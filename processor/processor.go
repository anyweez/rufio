package main

import (
	"time"
	"flag"
	"log"

	"github.com/luke-segars/rufio/shared/structs"
	proto "github.com/luke-segars/rufio/proto"
	"github.com/luke-segars/rufio/shared/queue"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/**
 * Processor is an instance that retrieves processing requests from a series of work queues,
 * reads data from existing collections, and updates or generates new objects to be stored in
 * others collections.
 *
 * It is composed of a primary function that invokes different COMPONENTS. Each component is
 * responsible for generating a particular type of output. Each component is made up of one or
 * more MODULES, which are responsible for consuming one or more types of input. Each module
 * is executed on each processed object in order to produce the final object. Modules follow
 * a map/reduce style workflow that involves retrieving a set of documents and then Join()'ing 
 * them in a meaningful way.
 */

// The URL for the Mongo database connection.
var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
// The name of the Mongo database.
var DB_NAME = flag.String("db", "league", "Name of the database where data should be read from / written to.")
// The URL and port for the Beanstalk address
var BEANSTALK_ADDRESS = flag.String("queue", "", "Address of Beanstalk queue.")
// The number of jobs that should be executed in parallel.
var NUM_JOBS = flag.Int("num_jobs", 10, "Number of jobs to run simultaneously.")

var components = make([]structs.IComponent, 0)

type JobRequest struct {
	Job		proto.ProcessedJobRequest
	Comp	structs.IComponent
	Listener *queue.QueueListener
}

func main() {
	flag.Parse()

	// Initialize shared (read-only) mappings
	mappings, err := structs.LoadMappings(*MONGO_CONNECTION_URL, *DB_NAME, []string{"game2rawgame", "game2summoner"})
	if err != nil {
		log.Fatal(err.Error())
	}

	// TODO: CreateMapping() for each mapping type

	workQueue := make(chan *JobRequest, *NUM_JOBS)
	for i := 0; i < *NUM_JOBS; i++ {
		go RunJob(workQueue)
	}

	// Kick off a component to listen to each job queue. The component will
	// read from the queue and handle each job that comes in.
	for _, component := range components {
		// Store the mappings that this component requires to function.
		component.StoreMappings(mappings)
		go RunComponent(component, workQueue)
	}

	// TODO: turn this into a WaitGroup.
	for {
		time.Sleep(60 * time.Second)
	}

}

// Each component consists of modules that are executed for each job. They read from
//   a single data source and annotate a single output object.
func RunComponent(comp structs.IComponent, requestQueue chan *JobRequest) {
	// Initialize connection to queue
	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{comp.GetQueueName()})
	if err != nil {
		log.Fatal(err.Error())
	}

	// For each job, kick off a goroutine that handles the rest of the processing.
	for job := range listener.Queue {
		requestQueue <- &JobRequest{
			Comp: comp,
			Job: job,
			Listener: &listener,
		}
	}
}

func RunJob(queue chan *JobRequest) {
	// TODO: Connect to database.
	session, err := mgo.Dial(*MONGO_CONNECTION_URL)
	if err != nil {
		log.Fatal(err.Error())
	}
	session.SetSocketTimeout(10 * time.Hour)
	db := session.DB(*DB_NAME)
	
	// For each job from queue, retrieve all docs and create new docs.
	for request := range queue {
		// Create a new document and check the database to see if one already exists.
		result := db.C(request.Comp.GetOutputCollection()).Find(bson.M{request.Comp.GetOutputKey(): *request.Job.TargetId})
		processed := request.Comp.GetProcessedType()

		// If a document already exists, use that instead.
		if cnt, _ := result.Count(); cnt > 0 {
			result.One(&processed)
		} 

		// Execute each module to update the processed object.
		for _, module := range request.Comp.GetModules() {
			// Get the mapping and fetch all ID's.
			mapping := module.GetMapping()
			ids := mapping.Get(*request.Job.TargetId)

			// Fetch all docs
			result = db.C(module.GetInputCollection()).Find(bson.M{module.GetInputKey(): ids})
			cnt, _ := result.Count()
			if cnt > 0 {
				docs := make([]structs.Document, cnt)
				result.All(&docs)
				// Join them together in a module-specific way.
				processed = module.Join(docs, processed, *request.Job.TargetId)
			}
		}

		// Update or insert the document after all modules have been processed.
		db.C(request.Comp.GetOutputCollection()).Upsert(bson.M{request.Comp.GetOutputKey(): *request.Job.TargetId}, processed)

		// Mark the job as completed.
		request.Listener.Finish(request.Job)
	}
}