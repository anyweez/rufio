package main

import (
	"github.com/luke-segars/rufio/shared/structs"
	"github.com/luke-segars/rufio/shared"
	"time"
)

/**
 * Read in from raw games table.
 */
type Raw2ProcessedGameModule struct {
	structs.Module
}

/**
 * All input documents contain a reference to `id` and should we merged into `existing`, then returned.
 */
func (rgm Raw2ProcessedGameModule) Join(docs []structs.Document, existing structs.Document, gameId int64) structs.Document {
	current := existing.(structs.ProcessedGameDocument)
	stats := make(map[int]structs.ProcessedPlayerStats)

	// Input documents are RawGameDoc's. Type assert and iterate through.
	for _, doc := range docs {
		response := doc.(structs.RawGameDocument).Response
		for _, game := range response.Games {
			// Skip all games that are not the target game.
			if int64(game.GameId) != gameId {
				continue
			}

			current.GameId = int(gameId)
			current.GameTimestamp = int64(game.CreateDate)
			current.GameDate = time.Unix(int64(game.CreateDate)/1000, 0).Format("2006-01-02")
			current.GameType = shared.GetGameType(game)

			// Put these into a separate dictionary so that they can be deduped before they're added
			// to the processed record.
			stats[response.SummonerId] = structs.ProcessedPlayerStats {
				SummonerId:       response.SummonerId,
				NumDeaths:        game.Stats.NumDeaths,
				MinionsKilled:    game.Stats.MinionsKilled,
				WardsPlaced:      game.Stats.WardPlaced,
				WardsCleared:     game.Stats.WardKilled,
				ChampionId:       game.ChampionId,

				// TODO: add other fields
			}
		}
	}

	// No wthat all have been dedup'd, add to the proper field.
	for _, stat := range stats {
		current.Stats = append(current.Stats, stat)
	}

	return current
}

type ProcessedLeague2ProcessedGameModule struct {
	structs.Module
}

func (rlm ProcessedLeague2ProcessedGameModule) Join(docs []structs.Document, existing structs.Document, gameId int64) structs.Document {
	current := existing.(structs.ProcessedGameDocument)

	// For each summoner we've got stats on, determine what their league ranking was at the time.
	for i, stats := range current.Stats {
		// Scan through all docs.
		for _, d := range docs {
			doc := d.(structs.ProcessedLeagueDocument)
			if doc.SummonerId != stats.SummonerId {
				continue
			}

			current.Stats[i].SummonerTier = doc.Current.Tier
			current.Stats[i].SummonerDivision = doc.Current.Division

			// TODO: should select appropriate league ranking based on when game was played		
		}
	}

	return existing
}