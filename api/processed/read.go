package processed

import (
	"gopkg.in/mgo.v2/bson"
	"shared/structs"
	"time"
)

/**
 * Get the ProcessedLeague for the summoner at a given time.
 */
func (api *ProcessedApi) GetLeagueAt(summonerId int, when time.Time) (structs.ProcessedLeagueRank, error) {
	collection := api.Session.DB("league").C("processed_leagues")

	record := structs.ProcessedLeague{}
	err := collection.Find(bson.M{"_id": summonerId}).One(&record)
	if err != nil {
		return structs.ProcessedLeagueRank{}, err
	}

	// TODO: I don't think this condition is fully correct. Still need to check lastKnown for certain conditions I *think*
	if len(record.Historical) == 0 {
		return record.Current, nil
	} else {
		selected := record.Current

		// Check over all historical records. If one of the historical records is more recent
		// than the requested timestamp, this is what we should return.
		for _, past := range record.Historical {
			if past.LastKnown.After(when) {
				selected = past
			}
		}

		return selected, nil
	}
}
