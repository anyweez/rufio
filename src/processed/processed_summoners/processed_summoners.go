package main

import (
	raw "api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	"gopkg.in/mgo.v2/bson"
	"log"
	"shared/queue"
	"shared/structs"
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
	raw_api, err := raw.NewRawApi(*MONGO_CONNECTION_URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	// This will be infinite unless `jc` is closed (which it currently isn't).
	for job := range listener.Queue {
		le := loglin.New("process_summoner", loglin.Fields{
			"target_id": *job.TargetId,
		})

		summoner := structs.ProcessedSummoner{
			SummonerId:  int(*job.TargetId),
			CurrentTier: "UNRANKED",
		}

		// TODO: set name and other summoner metadata

		//  Get game ID's.
		responses := raw_api.GetCompleteGamesBySummoner(summoner.SummonerId)
		for _, gr := range responses {
			for _, game := range gr.Games {
				summoner.CompleteGameIds = append(summoner.CompleteGameIds, game.GameId)
			}
		}

		summoner.IncompleteGameIds = append(summoner.IncompleteGameIds, raw_api.GetIncompleteGameIdsBySummoner(summoner.SummonerId)...)
		// Get the summoner's latest league rating.
		latest, err := raw_api.GetLatestLeague(summoner.SummonerId, "RANKED_SOLO_5x5")

		if err != nil {
			log.Println(err.Error())
		} else {
			division_str := "0"
			division := 0

			// Sort through all of the entries and find one of the requested participant.
			for _, entry := range latest.Entries {
				if entry.PlayerOrTeamId == latest.ParticipantId {
					summoner.CurrentTier = latest.Tier
					division_str = entry.Division
				}
			}

			// Convert the
			switch division_str {
			case "I":
				division = 1
				break
			case "II":
				division = 2
				break
			case "III":
				division = 3
				break
			case "IV":
				division = 4
				break
			case "V":
				division = 5
				break
			default:
				division = 0
				break
			}

			summoner.CurrentDivision = division
		}

		log.Println(fmt.Sprintf("Saving processed summoner #%d...", summoner.SummonerId))

		collection := raw_api.Session.DB("league").C("processed_summoners")
		collection.Upsert(bson.M{"_id": summoner.SummonerId}, summoner)
		log.Println("Done.")
		listener.Finish(job)

		le.Update(loglin.STATUS_COMPLETE, "", nil)
	}
}
