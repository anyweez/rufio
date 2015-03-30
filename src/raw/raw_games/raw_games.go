package main

import (
	"encoding/json"
	"flag"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net/http"
	fetcher "shared/data_fetcher"
	"shared/queue"
	"shared/structs"
)

// A few user-specified flags required for fetching summoner game data.
var API_KEY = flag.String("apikey", "", "Riot API key")
var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

const API_URL = "https://na.api.pvp.net/api/lol/na/v1.3/game/by-summoner/%d/recent?api_key=%s"

/**
 * This process instantiates a data fetcher that queries all
 */
func main() {
	flag.Parse()

	// Create the Mongo session.
	log.Println("Connecting to Mongo @ " + *MONGO_CONNECTION_URL)
	session, cerr := mgo.Dial(*MONGO_CONNECTION_URL)
	if cerr != nil {
		fmt.Println("Cannot connect to mongodb instance")
	}
	collection := session.DB("league").C("raw_games")
	log.Println("Done.")

	// Connect to beanstalk task queue to get summoner ID's.
	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{"retrieve_recent_games"})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create the data fetcher that's going to make all of the API requests
	// and store the data in StoreCollection (Mongo).
	requests := make(chan structs.FetchRequest)

	df := fetcher.NewDataFetcher(fetcher.DataFetcherConfig{
		RateLimit:  500,
		RatePeriod: 600,
		Requests:   requests,
		WithResponse: func(response *http.Response, req structs.FetchRequest) {
			body, _ := ioutil.ReadAll(response.Body)

			// Parse and store the response.
			gr := structs.NewGameResponse()
			json.Unmarshal(body, &gr.Response)

			// Store the response
			collection.Insert(gr)
			// Complete the job
			listener.Finish(req.Job)
		},
	})

	// Continuously retrieve jobs from the queue.
	for job := range listener.Queue {
		// TargetId for this job type are all summoner ID's.
		requests <- structs.FetchRequest{
			Job:   job,
			Queue: "retrieve_recent_games",
			Url:   fmt.Sprintf(API_URL, *job.TargetId, *API_KEY),
		}
	}

	//	close(requests)
	df.Close()
}
