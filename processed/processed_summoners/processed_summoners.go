package main

import (
	processed_api "github.com/luke-segars/loldata/api/processed"
	raw_api "github.com/luke-segars/loldata/api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	"gopkg.in/mgo.v2/bson"
	"log"
	proto "github.com/luke-segars/loldata/proto"
	"github.com/luke-segars/loldata/shared/queue"
	"github.com/luke-segars/loldata/shared/structs"
	"time"
)

var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

func main() {
	flag.Parse()

	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{"generate_processed_summoner"})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create a MongoDB session and save the data.
	raw, err := raw_api.NewRawApi(*MONGO_CONNECTION_URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	processed, err := processed_api.NewProcessedApi(*MONGO_CONNECTION_URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	// This will be infinite unless `jc` is closed (which it currently isn't).
	for job := range listener.Queue {
		le := loglin.New("process_summoner", loglin.Fields{
			"task":      proto.ProcessedJobRequest_GENERATE_PROCESSED_SUMMONER,
			"target_id": *job.TargetId,
		})

		summoner := structs.ProcessedSummoner{
			SummonerId:  int(*job.TargetId),
			CurrentTier: "UNRANKED",
		}

		le.Update(loglin.STATUS_OK, "Starting GetRawSummonerInfo", nil)
		// Set the summoner name (and potentially other metadata)
		rawSum, err := raw.GetRawSummonerInfo(summoner.SummonerId)
		le.Update(loglin.STATUS_OK, "Finishing GetRawSummonerInfo", nil)
		if err == nil {
			summoner.Name = rawSum.Name
		} else {
			le.Update(loglin.STATUS_WARNING, "No summoner name found.", nil)
		}

		//  Get game ID's.
		le.Update(loglin.STATUS_OK, "Starting GetCompleteGamesBySummoner", nil)
		responses := raw.GetCompleteGamesBySummoner(summoner.SummonerId)
		for _, gr := range responses {
			for _, game := range gr.Games {
				summoner.CompleteGameIds = append(summoner.CompleteGameIds, game.GameId)
			}
		}
		le.Update(loglin.STATUS_OK, "Finishing GetCompleteGamesBySummoner", nil)

		le.Update(loglin.STATUS_OK, "Starting GetIncompleteGameIdsBySummoner", nil)
		summoner.IncompleteGameIds = append(summoner.IncompleteGameIds, raw.GetIncompleteGameIdsBySummoner(summoner.SummonerId)...)
		le.Update(loglin.STATUS_OK, "Finishing GetIncompleteGameIdsBySummoner", nil)

		// Get the summoner's latest league rating.
		le.Update(loglin.STATUS_OK, "Starting GetLeagueAt", nil)
		league, err := processed.GetLeagueAt(summoner.SummonerId, time.Now())
		le.Update(loglin.STATUS_OK, "Finishing GetRawSummonerInfo", nil)

		if err != nil {
			le.Update(loglin.STATUS_WARNING, "No rank information available.", nil)
		} else {
			summoner.CurrentDivision = league.Division
			summoner.CurrentTier = league.Tier
		}

		log.Println(fmt.Sprintf("Saving processed summoner #%d...", summoner.SummonerId))

		le.Update(loglin.STATUS_OK, "Saving", nil)
		collection := raw.Session.DB("league").C("processed_summoners")
		collection.Upsert(bson.M{"_id": summoner.SummonerId}, summoner)
		listener.Finish(job)
		le.Update(loglin.STATUS_OK, "Save complete", nil)

		le.Update(loglin.STATUS_COMPLETE, "", loglin.Fields{
			"code": 200,
		})
	}
}
