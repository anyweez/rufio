package main

import (
	raw "api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"shared/queue"
	"shared/structs"
	"time"
	// "github.com/iwanbk/gobeanstalk"
	// "log"
	// gproto "code.google.com/p/goprotobuf/proto"
	// proto "proto"
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

	// Create a MongoDB session and save the data.
	log.Println("Connecting to Mongo @ " + *MONGO_CONNECTION_URL)
	session, cerr := mgo.Dial(*MONGO_CONNECTION_URL)
	if cerr != nil {
		fmt.Println("Cannot connect to mongodb instance")
		return
	}
	defer session.Close()

	collection := session.DB("league").C("processed_games")

	// This will be infinite unless `jc` is closed (which it currently isn't).
	for job := range listener.Queue {
		le := loglin.New("process_game", loglin.Fields{
			"target_id": *job.TargetId,
		})

		pg := structs.ProcessedGame{}
		pg.GameId = int(*job.TargetId)

		// Fetch all instances of raw games that have information about
		// this game ID and store them.
		gr := raw_api.GetPartialGames(pg.GameId)
		pps := make(map[int]structs.ProcessedPlayerStats)

		for _, response := range gr {
			for _, game := range response.Games {
				pg.GameTimestamp = int64(game.CreateDate)
				// Divide by one thousand since the value is in milliseconds.
				pg.GameDate = time.Unix(int64(game.CreateDate)/1000, 0).Format("2006-01-02")
				// Only do processing on the game that's being handled in this job.
				// Other games should be discarded.
				if game.GameId == pg.GameId {
					// TODO: instead of getting 'latest', should get 'closest to timestamp X (but not after)'.
					// Current approach works fine unless we're running a backfill.
					latestLeague, lerr := raw_api.GetLatestLeague(response.SummonerId, "RANKED_SOLO_5x5")
					tier := "UNRANKED"
					division_str := "0"
					division := 0

					if lerr == nil {
						// Sort through all of the entries and find one of the requested participant.
						for _, entry := range latestLeague.Entries {
							if entry.PlayerOrTeamId == latestLeague.ParticipantId {
								tier = latestLeague.Tier
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
		}

		// Copy one PPS entry per summoner into the processed game file.
		for _, v := range pps {
			pg.Stats = append(pg.Stats, v)
		}

		log.Println(fmt.Sprintf("Saving processed game #%d...", pg.GameId))
		collection.Upsert(bson.M{"_id": pg.GameId}, pg)
		log.Println("Done.")
		listener.Finish(job)

		le.Update(loglin.STATUS_COMPLETE, "", nil)
	}
}
