package main

import (
	raw "api/raw"
	"flag"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"log"
	"shared/queue"
	"shared/structs"
)

var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

func main() {
	flag.Parse()
	// TODO: replace this with pulling something from a live queue.
	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{"generate_processed_summoner"})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create a MongoDB session and save the data.
	log.Println("Connecting to Mongo @ " + *MONGO_CONNECTION_URL)
	session, cerr := mgo.Dial(*MONGO_CONNECTION_URL)
	if cerr != nil {
		fmt.Println("Cannot connect to mongodb instance")
		return
	}
	collection := session.DB("league").C("processed_summoners")
	defer session.Close()

	// This will be infinite unless `jc` is closed (which it currently isn't).
	for job := range listener.Queue {
		summoner := structs.ProcessedSummoner{
			SummonerId: int(*job.TargetId),
		}

		// TODO: set name and other summoner metadata

		//  Get game ID's.
		responses := raw.GetCompleteGamesBySummoner(summoner.SummonerId)
		for _, gr := range responses {
			for _, game := range gr.Games {
				summoner.CompleteGameIds = append(summoner.CompleteGameIds, game.GameId)
			}
		}

		summoner.IncompleteGameIds = append(summoner.IncompleteGameIds, raw.GetIncompleteGameIdsBySummoner(summoner.SummonerId)...)

		// Get the summoner's latest league rating.
		latest, err := raw.GetLatestLeague(summoner.SummonerId, "RANKED_SOLO_5x5")
		if err != nil {
			log.Println(err.Error())
		} else {
			tier := "UNKNOWN"
			division_str := "0"
			division := 0

			// Sort through all of the entries and find one of the requested participant.
			for _, entry := range latest.Entries {
				if entry.PlayerOrTeamId == latest.ParticipantId {
					tier = latest.Tier
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

			summoner.CurrentTier = tier
			summoner.CurrentDivision = division
		}

		log.Println(fmt.Sprintf("Saving processed summoner #%d...", summoner.SummonerId))
		collection.Insert(summoner)
		log.Println("Done.")
		listener.Finish(job)
	}
}
