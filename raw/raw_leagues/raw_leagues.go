package main

import (
	"encoding/json"
	"flag"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net/http"
	fetcher "github.com/luke-segars/loldata/shared/data_fetcher"
	"github.com/luke-segars/loldata/shared/queue"
	"github.com/luke-segars/loldata/shared/structs"
)

// A few user-specified flags required for fetching summoner game data.
var API_KEY = flag.String("apikey", "", "Riot API key")
var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

const API_URL = "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/%d?api_key=%s"

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
	collection := session.DB("league").C("raw_leagues")
	log.Println("Done.")

	// Load in summoner ID's and start generating URL's.
	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{"retrieve_recent_league"})
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
			gr := structs.NewLeagueResponse()
			json.Unmarshal(body, &gr.Response)

			// Store the response
			collection.Insert(gr)
			listener.Finish(req.Job)
		},
	})

	for job := range listener.Queue {
		// TargetId for this job type are all summoner ID's.
		requests <- structs.FetchRequest{
			Job:   job,
			Queue: "retrieve_recent_league",
			Url:   fmt.Sprintf(API_URL, *job.TargetId, *API_KEY),
		}
	}

	//	close(requests)
	df.Close()
}
