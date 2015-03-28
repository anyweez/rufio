package structs

import (
	"time"
)

type RequestMetadata struct {
	RequestTime time.Time
	SummonerId  int
}

// Master wrapper for game API + metadata.
type GameResponseWrapper struct {
	Metadata RequestMetadata
	Response GameResponse
}

// Wrapper for game API.
type GameResponse struct {
	SummonerId int
	Games      []GameResponseGame
}

// Game object from within a GameResponse.
type GameResponseGame struct {
	GameId        int
	Invalid       bool
	GameMode      string
	SubType       string
	MapId         int
	TeamId        int
	ChampionId    int
	Spell1        int
	Spell2        int
	Level         int
	IpEarned      int
	CreateDate    int // reported in milliseconds
	FellowPlayers []GameResponsePlayer
	Stats         GameResponseStats
}

type GameResponsePlayer struct {
	SummonerId int
	TeamId     int
	ChampionId int
}

type GameResponseStats struct {
	Level                           int
	GoldEarned                      int
	NumDeaths                       int
	MinionsKilled                   int
	ChampionsKilled                 int
	GoldSpent                       int
	TotalDamageDealt                int
	TotalDamageTaken                int
	KillingSprees                   int
	LargestKillingSpree             int
	Team                            int
	Win                             bool
	NeutralMinionsKilled            int
	LargestMultiKill                int
	PhysicalDamageDealtPlayer       int
	MagicDamageDealtPlayer          int
	PhysicalDamageTaken             int
	MagicDamageTaken                int
	TimePlayed                      int
	TotalHeal                       int
	TotalUnitsHealed                int
	Assists                         int
	Item0                           int
	Item1                           int
	Item2                           int
	Item3                           int
	Item4                           int
	Item5                           int
	Item6                           int
	SightWardsBought                int
	MagicDamageDealtToChampions     int
	PhysicalDamageDealtToChampions  int
	TotalDamageDealtToChampions     int
	TrueDamageTaken                 int
	WardKilled                      int
	WardPlaced                      int
	NeutralMinionsKilledEnemyJungle int
	NeutralMinionsKilledYourJungle  int
	TotalTimeCrowdControlDealt      int
	PlayerRole                      int
	PlayerPosition                  int
}

func NewGameResponse() GameResponseWrapper {
	grw := GameResponseWrapper{}
	grw.Metadata.RequestTime = time.Now()

	return grw
}
