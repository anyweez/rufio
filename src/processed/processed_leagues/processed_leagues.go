package main

import (
	processedapi "api/processed"
	rawapi "api/raw"
	"flag"
	"fmt"
	"github.com/luke-segars/loglin"
	"log"
	shared "shared"
	"shared/structs"
	"strconv"
	"time"
)

var MONGO_CONNECTION = flag.String("mongodb", "localhost", "A connection string identifying a MongoDB instance.")
var TARGET_IDS = flag.String("target_ids", "input/summoners", "A list of summoner ID's that should have records generated.")

type ParsedLeagueInfo struct {
	SummonerId int
	Tier       string
	Division   int
	LastKnown  time.Time
}

/**
 * This program runs over all raw_league events and generates current as well as past league
 * events, keyed by summoner ID.
 *
 * Note that it expects that log entries for an individual summoner are retrieved in chronological
 * orer. In order to keep this process simple it's currently ignoring historical records that occur
 * before the "current" entry for a summoner. This means that the only way to backfill historical
 * data is to remove more recent data and regenerate the entire record holistically.
 *
 * Note that for my current use cases this is a perfectly adequate assumption but may bite me later.
 */

/**
 * Extracts the summoner id, tier, division, and timestamp from this object.
 */
func extractLeagueInfo(record structs.LeagueResponseWrapper) []ParsedLeagueInfo {
	pli := make([]ParsedLeagueInfo, 0)

	// Find the correct LeagueResponseTier (for the right summoner, solo queue).
	ltr := structs.LeagueResponseTier{}
	for _, tiers := range record.Response {
		for _, tier := range tiers {
			if tier.Queue == "RANKED_SOLO_5x5" {
				ltr = tier
			}
		}
	}

	// If an LTR was found for solo queue, iterate through all of those entries and
	// generate ParsedLeagueInfo objects.
	if len(ltr.Name) > 0 {
		for _, entry := range ltr.Entries {
			summonerId, _ := strconv.ParseInt(entry.PlayerOrTeamId, 10, 32)
			division := 0

			switch entry.Division {
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
			}

			pli = append(pli, ParsedLeagueInfo{
				SummonerId: int(summonerId),
				Tier:       ltr.Tier,
				Division:   division,
				LastKnown:  record.Metadata.RequestTime,
			})
		}
	}

	// Return the list of all ParsedLeagueInfo objects.
	return pli
}

func main() {
	flag.Parse()
	fmt.Println("Reading summoner ID's.")
	summ, err := shared.LoadIds(*TARGET_IDS)
	if err != nil {
		log.Fatal("Couldn't load summoner ID file: " + err.Error())
	}

	// Load summoner list and convert it into a map for O(1) lookups.
	summoners := make(map[int]bool)
	for _, id := range summ {
		fmt.Println(id)
		summoners[id] = true
	}

	fmt.Println("Connecting to database @ " + *MONGO_CONNECTION)
	le := loglin.New("processed_leagues", nil)
	// TODO: load initial history from the table.
	history := make(map[int]structs.ProcessedLeague)

	// Connect to MongoDB.
	api, err := rawapi.NewRawApi(*MONGO_CONNECTION)
	if err != nil {
		log.Fatal("Couldn't initialize raw data API: " + err.Error())
	}
	// Increase timeout since we're pumping a ton of data through.
	// TODO: can all of these queries be converted to a single bulk request so that
	// timeouts don't occur?
	api.Session.SetSocketTimeout(1 * time.Hour)

	le.Update(loglin.STATUS_OK, "Connected to MongoDB", nil)

	// Get all raw records for all users.
	collection := api.Session.DB("league").C("raw_leagues")

	// TODO: order by Metadata.RequestTime
	iter := collection.Find(nil).Iter()
	defer iter.Close()

	record := structs.LeagueResponseWrapper{}
	for iter.Next(&record) {
		// For each record, check to see if there's an existing user history. if so,
		// check to see if the divion and tier are the same.
		leagueInfo := extractLeagueInfo(record)
		for _, li := range leagueInfo {
			// Check to see if this is an eligible summoner (if it's included on the list).
			if _, eligible := summoners[li.SummonerId]; eligible {
				// Check to see if there's an existing user history.
				league, exists := history[li.SummonerId]

				// If the object already exists, we need to see whether updating the current object is
				// required.
				if exists {
					// If the new league info is different than current info, bump "current" to
					// historical and create a new "current."
					if li.LastKnown.After(league.Current.LastKnown) && (league.Current.Tier != li.Tier || league.Current.Division != li.Division) {
						le.Update(loglin.STATUS_OK, "League record for existing summoner.", loglin.Fields{
							"op":         "update",
							"summonerid": li.SummonerId,
							"tier":       li.Tier,
							"division":   li.Division,
						})

						league.Historical = append(league.Historical, league.Current)
						league.Current.LastKnown = li.LastKnown
						league.Current.Tier = li.Tier
						league.Current.Division = li.Division
					}
					// If it doesn't already exist, we should create it and set "current" to whatever
					// values we have here.
				} else {
					le.Update(loglin.STATUS_OK, "League record for new summoner.", loglin.Fields{
						"op":         "new",
						"summonerid": li.SummonerId,
						"tier":       li.Tier,
						"division":   li.Division,
					})

					history[li.SummonerId] = structs.ProcessedLeague{
						SummonerId: li.SummonerId,
						LastUpdate: time.Now(),
						Current: structs.ProcessedLeagueRank{
							Tier:     li.Tier,
							Division: li.Division,
						},
					}
				}
			}
		}
	}

	le.Update(loglin.STATUS_OK, "Writing output", nil)

	// Write the processed league object to storage.
	//	collection = api.Session.DB("league").C("processed_leagues")
	processed, err := processedapi.NewProcessedApi(*MONGO_CONNECTION)

	for _, league := range history {
		err := processed.StoreLeague(league)
		if err != nil {
			// le.Update(loglin.STATUS_WARNING, err.Error(), loglin.Fields{
			// 	"summonerId": summonerId,
			// })
		}
	}
}
