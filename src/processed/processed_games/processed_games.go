package main

import (
	processed "api/processed"
	raw "api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	proto "proto"
	shared "shared"
	"shared/queue"
	"shared/structs"
	"time"
)

var MONGO_CONNECTION_URL = flag.String("mongodb", "localhost", "The URL that mgo should use to connect to Mongo.")
var BEANSTALK_ADDRESS = flag.String("queue", "localhost:11300", "[host:port] The address of the beanstalk queue to pull jobs from.")

func main() {
	flag.Parse()

	listener, err := queue.NewQueueListener(*BEANSTALK_ADDRESS, []string{"generate_processed_game"})
	if err != nil {
		log.Fatal(err.Error())
	}

	raw_api, err := raw.NewRawApi(*MONGO_CONNECTION_URL)
	if err != nil {
		log.Fatal(err.Error())
	}

	processed_api, err := processed.NewProcessedApi(*MONGO_CONNECTION_URL)
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
	defer session.Close()

	// This will be infinite unless `jc` is closed (which it currently isn't).
	for job := range listener.Queue {
		le := loglin.New("process_game", loglin.Fields{
			"target_id": *job.TargetId,
			"task":      proto.ProcessedJobRequest_GENERATE_PROCESSED_GAME,
		})

		pg := structs.ProcessedGame{}
		pg.GameId = int(*job.TargetId)

		// Fetch all instances of raw games that have information about
		// this game ID and store them.
		gr, err := raw_api.GetPartialGames(pg.GameId)
		if err != nil {
			// If we can't fetch data, abandon.
			le.Update(loglin.STATUS_ERROR, err.Error(), nil)
			listener.Finish(job)
			continue
		}

		// One pps container per job (game)
		pps := make(map[int]structs.ProcessedPlayerStats)

		for _, response := range gr {
			for _, game := range response.Games {
				// GetPartialGames currently returns all games within a record if one of them has the desired gid.
				// We need to filter out most at the application level until I get the mongo query refined.
				if game.GameId != pg.GameId {
					continue
				}

				pg.GameTimestamp = int64(game.CreateDate)
				// Divide by one thousand since the value is in milliseconds.
				pg.GameDate = time.Unix(int64(game.CreateDate)/1000, 0).Format("2006-01-02")
				// Get game type
				pg.GameType = shared.GetGameType(game)

				latestLeague, lerr := processed_api.GetLeagueAt(response.SummonerId, time.Unix(int64(game.CreateDate)/1000, 0))
				tier := "UNRANKED"
				division := 0

				// If no errors, set the tier and division
				if lerr == nil {
					// If Tier is set, copy it over.
					if len(latestLeague.Tier) > 0 {
						tier = latestLeague.Tier
					}	
					// If Division is set, copy it over.
					if latestLeague.Division > 0 {
						division = latestLeague.Division
					}
				} else {
					le.Update(loglin.STATUS_WARNING, lerr.Error(), nil)
				}

				// This GameRecord has enough information to populate one user's
				// ProcessedPlayerStats. Generate that object, add it to the game,
				// and look for others.
				pps[response.SummonerId] = structs.ProcessedPlayerStats{
					SummonerId:       response.SummonerId,
					SummonerTier:     tier,
					SummonerDivision: division,
					NumDeaths:        game.Stats.NumDeaths,
					MinionsKilled:    game.Stats.MinionsKilled,
					WardsPlaced:      game.Stats.WardPlaced,
					WardsCleared:     game.Stats.WardKilled,
				}
			}
		}

		// Copy one PPS entry per summoner into the processed game file.
		for _, v := range pps {
			pg.Stats = append(pg.Stats, v)
		}

		le.Update(loglin.STATUS_OK, "Storing", nil)

		collection := session.DB("league").C("processed_games")
		_, err = collection.Upsert(bson.M{"_id": pg.GameId}, pg)
		if err != nil {
			le.Update(loglin.STATUS_ERROR, err.Error(), nil)
		} else {
			log.Println("Done.")
		}

		// Mark job as complete.
		listener.Finish(job)

		le.Update(loglin.STATUS_COMPLETE, "", loglin.Fields{
			"code": 200,
		})
	}
}
