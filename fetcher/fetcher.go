package main

import (
	"flag"
	"time"
	"log"
	"net/http"
	"github.com/luke-segars/loglin"
	mgo "gopkg.in/mgo.v2"
	proto "github.com/luke-segars/rufio/proto"
	queue "github.com/luke-segars/rufio/shared/queue"
	structs "github.com/luke-segars/rufio/shared/structs"
)

type FetcherConfig struct {
	FetchType		string
	FetchParser		func(*http.Response) (structs.RawResponseWrapper, error)
	TubeName 		string // Beanstalk queue name

	DatabaseName 	string
	CollectionName 	string

	BuildUrl		func(proto.ProcessedJobRequest, string) string
}

var configs []FetcherConfig

// A few user-specified flags required for fetching summoner game data.
var API_KEY = flag.String("apikey", "", "Riot API key")
var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

func init() {
	// Load up
	configs = make([]FetcherConfig, 0)
}

func main() {
	flag.Parse()
	// Current API rate limit is 500 per 10 minutes, which means 50 per minute, or one every 6/5 seconds.
	rate := NewRateLimiter((6.0 / 5) * 1000 * time.Millisecond)

	for _, config := range configs {
		go Fetcher(config, rate)
	}

	// TODO: replace this with a goroutine-based block of some sort. Other goroutines will handle it from here.
	for {
		time.Sleep(60 * time.Second)
	}
}

/**
 * A new fetcher is spawned for each FetchType that this process is responsible for. The same flow should be
 * usable for anything that queries the Riot API; you just need to create a new FetcherConfig (see fetchgame.go
 * for an example).
 */
func Fetcher(config FetcherConfig, rate chan bool) {
	// Connect to database
	session, cerr := mgo.Dial(*MONGO_CONNECTION_URL)
	if cerr != nil {
		log.Fatal("Cannot connect to MongoDB instance: " + cerr.Error())
		return
	}
	collection := session.DB(config.DatabaseName).C(config.CollectionName)

	// Connect to beanstalk task queue to get summoner ID's.
	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{config.TubeName})
	if err != nil {
		log.Fatal("Cannot connect to beanstalkd instance: " + err.Error())
		return
	}

	// Continuously retrieve jobs from the queue.
	for job := range listener.Queue {
		<- rate
	
		le := loglin.New(config.CollectionName, loglin.Fields{
			"target_id": *job.TargetId,
			"task":      *job.Type,
		})

		// TargetId for this job type are all summoner ID's.
		req := structs.FetchRequest{
			Job:   job,
			Queue: config.TubeName,
			Url:   config.BuildUrl(job, *API_KEY), //fmt.Sprintf(API_URL, *job.TargetId, *API_KEY),
		}

		// Make the request and store the outcome in the provided collection.
		go request(req, le, func(response *http.Response, req structs.FetchRequest) {
			// Parse the response and insert it into the database.
			parsed, err := config.FetchParser(response)
			if err != nil {
				le.Update(loglin.STATUS_ERROR, "Cannot parse: " + err.Error(), nil)
				return
			}

			collection.Insert(parsed)
			
			// Complete the job
			listener.Finish(req.Job)
			le.Update(loglin.STATUS_COMPLETE, "Job deleted from queue.", nil)
		})
	}
}

/**
 * Make an HTTP GET request and handle the response using the withResponse function.
 */
func request(request structs.FetchRequest, le loglin.LogEvent, withResponse func(response *http.Response, req structs.FetchRequest)) {
	// Check to make sure there's actually a job coming through.
	if request.Job.TargetId == nil || request.Job.Type == nil {
		return
	}

	resp, err := http.Get(request.Url)

	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			le.Update(loglin.STATUS_ERROR, err.Error(), loglin.Fields{
				"code":  -1,
				"stage": "retrieval",
			})
		} else {
			le.Update(loglin.STATUS_ERROR, "Retrieval error; abandoning for now", loglin.Fields{
				"code":  resp.StatusCode,
				"stage": "retrieval",
			})
		}
	// On success, save
	} else {
		defer resp.Body.Close()

		le.Update(loglin.STATUS_OK, "", loglin.Fields{
			"code":  resp.StatusCode,
			"stage": "retrieval",
		})

		withResponse(resp, request)

		le.Update(loglin.STATUS_COMPLETE, "", loglin.Fields{
			"code":  resp.StatusCode,
			"stage": "storage",
		})
	}
}