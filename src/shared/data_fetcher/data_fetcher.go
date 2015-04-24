package data_fetcher

import (
	//	"encoding/json"
	"github.com/luke-segars/loglin"
	// mgo "gopkg.in/mgo.v2"
	// "io/ioutil"
	// "math"
	"net/http"
	"shared/structs"
	"time"
)

type DataFetcherConfig struct {
	Requests     chan structs.FetchRequest
	WithResponse func(*http.Response, structs.FetchRequest)
	RateLimit    int // number of requests per period
	RatePeriod   int // length of period (in seconds)
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
	secondsPerRequest := float64(df.Config.RatePeriod) / float64(df.Config.RateLimit)
	for {
		time.Sleep(time.Duration(secondsPerRequest) * time.Second)
		df.Ready <- true
	}
}

/*
func currentBucket() int {
	return int(math.Floor(float64(time.Now().Second()) / 13))
}
*/

func run_fetcher(df *DataFetcher) {
	// Inifinite loop to continuously fetch data (restricted by rate limiter).
	for request := range df.Config.Requests {
		<-df.Ready
		// Fetch recent game data.
		go fetch(request, df)
	}

}

/**
 * Fetch the data from the provided URL and (eventually) store it.
 */
func fetch(req structs.FetchRequest, df *DataFetcher) {
	// Check to make sure there's actually a job coming through.
	if req.Job.TargetId == nil || req.Job.Type == nil {
		return
	}

	le := loglin.New(req.Queue, loglin.Fields{
		"target_id": *req.Job.TargetId,
		"task":      *req.Job.Type,
	})
	resp, err := http.Get(req.Url)

	if err != nil || resp.StatusCode == 404 {
		if err != nil {
			le.Update(loglin.STATUS_ERROR, err.Error(), loglin.Fields{
				"code":  -1,
				"stage": "retrieval",
			})
		} else {
			le.Update(loglin.STATUS_ERROR, "", loglin.Fields{
				"code":  resp.StatusCode,
				"stage": "retrieval",
			})
		}
	} else {
		defer resp.Body.Close()

		le.Update(loglin.STATUS_OK, "", loglin.Fields{
			"code":  resp.StatusCode,
			"stage": "retrieval",
		})

		df.Config.WithResponse(resp, req)

		le.Update(loglin.STATUS_COMPLETE, "", loglin.Fields{
			"code":  resp.StatusCode,
			"stage": "storage",
		})
	}
}
