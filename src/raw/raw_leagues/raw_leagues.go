package main

import (
	"encoding/json"
	"flag"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"net/http"
	shared "shared"
	fetcher "shared/data_fetcher"
	"shared/structs"
)

// A few user-specified flags required for fetching summoner game data.
var API_KEY = flag.String("apikey", "", "Riot API key")
var CHAMPION_LIST = flag.String("summoners", "champions", "List of summoner ID's")
var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")

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

	// Create the data fetcher that's going to make all of the API requests
	// and store the data in StoreCollection (Mongo).
	urlChannel := make(chan string)

	df := fetcher.NewDataFetcher(fetcher.DataFetcherConfig{
		RateLimit: 2,
		Urls:      urlChannel,
		WithResponse: func(response *http.Response, url string) {
			body, _ := ioutil.ReadAll(response.Body)

			// Parse and store the response.
			gr := structs.NewLeagueResponse()
			json.Unmarshal(body, &gr.Response)
			fmt.Println(fmt.Sprintf("%d: %s", len(gr.Response), url))

			// Store the response
			collection.Insert(gr)
		},
	})

	// Load in summoner ID's and start generating URL's.
	summoner_ids, serr := shared.LoadSummonerIds(*CHAMPION_LIST)
	if serr != nil {
		fmt.Println(serr.Error())
		return
	}

	for _, summoner_id := range summoner_ids {
		urlChannel <- fmt.Sprintf(API_URL, summoner_id, *API_KEY)
	}

	close(urlChannel)
	df.Close()
}
