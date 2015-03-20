package data_fetcher

import (
	//	"encoding/json"
	"fmt"
	// mgo "gopkg.in/mgo.v2"
	// "io/ioutil"
	// "math"
	"net/http"
	"time"
)

const (
	RIOT_API_LIMIT_BUCKET_SIZE = 10 // Riot limits to X queries per 10 seconds.
)

type DataFetcherConfig struct {
	Urls         chan string
	WithResponse func(*http.Response, string)
	RateLimit    int
}

type DataFetcher struct {
	Ready  chan bool
	Config DataFetcherConfig
}

func NewDataFetcher(dfc DataFetcherConfig) DataFetcher {
	df := DataFetcher{}
	df.Initialize(dfc)

	return df
}

func (df *DataFetcher) Initialize(dfc DataFetcherConfig) {
	df.Ready = make(chan bool)
	df.Config = dfc

	// Start the rate limited and fetcher goroutines
	go manage_rate(df)
	go run_fetcher(df)
}

/**
 * TODO: make this more bulletproof. Should return when the URL queue is empty.
 */
func (df *DataFetcher) Close() {
	time.Sleep(3 * time.Second)
}

/**
 * A separate goroutine that pumps values onto the `ready` channel when the rate limit
 * hasn't been exceeded. It should be impossible to get back more than RATE_LIMIT values
 * from the channel in a single second.
 */
func manage_rate(df *DataFetcher) {
	// 10 is the "bucket size" that Riot uses to define it's API.
	secondsPerRequest := float64(RIOT_API_LIMIT_BUCKET_SIZE) / float64(df.Config.RateLimit)
	for {
		time.Sleep(time.Duration(secondsPerRequest) * time.Second)
		df.Ready <- true
	}
	/*
		time.AfterFunc(1*time.Second, func() {
			df.Ready <- true
		})
		current_bucket := currentBucket()
		count := 0

		for {
			// If it's a new second, restart the counter.
			if currentBucket() != current_bucket {
				current_bucket = currentBucket()
				count = 0
			}

			// Check to see if the rate limit is still valid.
			if count < df.Config.RateLimit {
				df.Ready <- true
				count += 1
			}

			time.Sleep(10 * time.Millisecond)
		}
	*/
}

/*
func currentBucket() int {
	return int(math.Floor(float64(time.Now().Second()) / 13))
}
*/

func run_fetcher(df *DataFetcher) {
	// Inifinite loop to continuously fetch data (restricted by rate limiter).
	for url := range df.Config.Urls {
		<-df.Ready
		// Fetch recent game data.
		go fetch(url, df)
	}

}

/**
 * Fetch the data from the provided URL and (eventually) store it.
 */
func fetch(url string, df *DataFetcher) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		defer resp.Body.Close()
		df.Config.WithResponse(resp, url)
	}
}
