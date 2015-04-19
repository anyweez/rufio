package structs

import (
	"time"
)

type ProcessedLeague struct {
	SummonerId int `bson:"_id"`
	LastUpdate time.Time

	Current    ProcessedLeagueRank
	Historical []ProcessedLeagueRank
}

type ProcessedLeagueRank struct {
	// The approximate time that the user was promoted to this rank.
	LastKnown time.Time

	Tier     string
	Division int
}
