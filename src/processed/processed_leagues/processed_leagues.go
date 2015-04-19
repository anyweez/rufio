package main

import (
	rawapi "api/raw"
	"flag"
	"github.com/luke-segars/loglin"
	"gopkg.in/mgo.v2/bson"
	"log"
	"shared/structs"
	"strconv"
	"time"
)

var MONGO_CONNECTION = flag.String("mongodb", "localhost", "A connection string identifying a MongoDB instance.")

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
	le := loglin.New("processed_leagues", nil)
	history := make(map[int]structs.ProcessedLeague)

	// Connect to MongoDB.
	api, err := rawapi.NewRawApi(*MONGO_CONNECTION)
	if err != nil {
		log.Fatal("Couldn't initialize raw data API: " + err.Error())
	}
	le.Update(loglin.STATUS_OK, "Connected to MongoDB", nil)

	// Get all raw records for all users.
	collection := api.Session.DB("league").C("raw_leagues")
	// TODO: order by Metadata.RequestTime
	iter := collection.Find(nil).Iter()
	defer iter.Close()
	le.Update(loglin.STATUS_OK, "Iterator retrieved", nil)

	record := structs.LeagueResponseWrapper{}
	for iter.Next(&record) {
		le.Update(loglin.STATUS_OK, "started", nil)
		// For each record, check to see if there's an existing user history. if so,
		// check to see if the divion and tier are the same.
		leagueInfo := extractLeagueInfo(record)

		le.Update(loglin.STATUS_OK, "League info extracted", nil)
		for _, li := range leagueInfo {
			// Check to see if there's an existing user history.
			league, exists := history[li.SummonerId]

			// If the object already exists, we need to see whether updating the current object is
			// required.
			if exists {
				// If the new league info is different than current info, bump "current" to
				// historical and create a new "current."
				if li.LastKnown.After(league.Current.PromotionTime) && (league.Current.Tier != li.Tier || league.Current.Division != li.Division) {
					le.Update(loglin.STATUS_OK, "Updating new league status", loglin.Fields{
						"op":         "new",
						"summonerid": li.SummonerId,
						"tier":       li.Tier,
						"division":   li.Division,
					})

					league.Historical = append(league.Historical, league.Current)
					league.Current.PromotionTime = li.LastKnown
					league.Current.Tier = li.Tier
					league.Current.Division = li.Division
				}
				// If it doesn't already exist, we should create it and set "current" to whatever
				// values we have here.
			} else {
				le.Update(loglin.STATUS_OK, "League record for new summoner.", loglin.Fields{
					"op":         "update",
					"summonerid": li.SummonerId,
					"tier":       li.Tier,
					"division":   li.Division,
				})

				history[li.SummonerId] = structs.ProcessedLeague{
					SummonerId: li.SummonerId,
					LastUpdate: time.Now(),
					Current: structs.ProcessedLeagueRank{
						PromotionTime: li.LastKnown,
						Tier:          li.Tier,
						Division:      li.Division,
					},
				}
			}
		}
	}

	// Write the processed league object to storage.
	collection = api.Session.DB("league").C("processed_leagues")
	for summonerId, league := range history {
		collection.Upsert(bson.M{"summonerid": summonerId}, league)
	}
}
