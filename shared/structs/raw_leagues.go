package structs

import (
	"time"
)

type ResponseMetadata struct {
	RequestTime time.Time
}

type LeagueResponseWrapper struct {
	Metadata ResponseMetadata
	Response map[string][]LeagueResponseTier
}

type LeagueResponseTier struct {
	Name          string
	Tier          string
	Queue         string
	Entries       []LeagueResponseEntry
	ParticipantId string
}

type LeagueResponseEntry struct {
	PlayerOrTeamId   string
	PlayerOrTeamName string
	Division         string
	LeaguePoints     int
	Wins             int
	Losses           int
	IsHotStreak      int
	IsVeteran        int
	IsFreshBlood     int
	IsInactive       int
}

func NewLeagueResponse() LeagueResponseWrapper {
	grw := LeagueResponseWrapper{}
	grw.Metadata.RequestTime = time.Now()

	return grw
}
