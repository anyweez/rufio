package main

import (
	processed_api "api/processed"
	raw_api "api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	"gopkg.in/mgo.v2/bson"
	"log"
	proto "proto"
	"shared/queue"
	"shared/structs"
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

		// Set the summoner name (and potentially other metadata)
		rawSum, err := raw.GetRawSummonerInfo(summoner.SummonerId)
		if err == nil {
			summoner.Name = rawSum.Name
			fmt.Println("Name: " + summoner.Name)
			fmt.Println("%+v", rawSum)
		} else {
			le.Update(loglin.STATUS_WARNING, "No summoner name found.", nil)
		}

		//  Get game ID's.
		responses := raw.GetCompleteGamesBySummoner(summoner.SummonerId)
		for _, gr := range responses {
			for _, game := range gr.Games {
				summoner.CompleteGameIds = append(summoner.CompleteGameIds, game.GameId)
			}
		}

		summoner.IncompleteGameIds = append(summoner.IncompleteGameIds, raw.GetIncompleteGameIdsBySummoner(summoner.SummonerId)...)
		// Get the summoner's latest league rating.
		league, err := processed.GetLeagueAt(summoner.SummonerId, time.Now())

		if err != nil {
			le.Update(loglin.STATUS_WARNING, "No rank information available.", nil)
		} else {
			summoner.CurrentDivision = league.Division
			summoner.CurrentTier = league.Tier
		}

		log.Println(fmt.Sprintf("Saving processed summoner #%d...", summoner.SummonerId))

		collection := raw.Session.DB("league").C("processed_summoners")
		collection.Upsert(bson.M{"_id": summoner.SummonerId}, summoner)
		log.Println("Done.")
		listener.Finish(job)

		le.Update(loglin.STATUS_COMPLETE, "", loglin.Fields{
			"code": 200,
		})
	}
}
