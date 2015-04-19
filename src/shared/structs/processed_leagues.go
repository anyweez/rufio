package structs

import (
	"time"
)

type ProcessedLeague struct {
	SummonerId int
	LastUpdate time.Time

	Current    ProcessedLeagueRank
	Historical []ProcessedLeagueRank
}

type ProcessedLeagueRank struct {
	// The approximate time that the user was promoted to this rank.
	PromotionTime time.Time

	Tier     string
	Division int
}
