package structs

import (
	"time"
)


// Metadata associated with the request used to generate the raw docs.
type DocumentMetadata struct {
	CreationTime 	time.Time
	TargetId  		int
	// TODO: fetcher needs to populate this.
	DocumentId 		int64
}

/**
 * RawGameDocument's store responses and related metadata for raw game responses from
 * the Riot API.
 */

type RawGameDocument struct {
	Metadata DocumentMetadata
	Response GameResponse
}

func (rgd RawGameDocument) GetId() int {
	return rgd.Metadata.TargetId
}

func (rgd RawGameDocument) GetCreationTime() time.Time {
	return rgd.Metadata.CreationTime
}

func (rgd RawGameDocument) GetType() DocType {
	return RawGameDoc
}

/**
 * RawLeagueDocument
 */


/**
 * RawSummonerDocument
 */

/**
 * ProcessedGameDocument
 */
type ProcessedGameDocument struct {
	Metadata 		DocumentMetadata
	GameId        	int `bson:"_id"`
	GameTimestamp 	int64
	// String representation of the above timestamp in "YYYY-MM-DD" format.
	GameDate 		string
	// Game type (see above)
	GameType 		ProcessedGameType
	Stats 			[]ProcessedPlayerStats
}

func (pgd ProcessedGameDocument) GetId() int {
	return pgd.GameId
}

func (pgd ProcessedGameDocument) GetCreationTime() time.Time {
	return pgd.Metadata.CreationTime
}

func (pgd ProcessedGameDocument) GetType() DocType {
	return ProcessedGameDoc
}

type ProcessedPlayerStats struct {
	SummonerId int
	// "BRONZE", etc
	SummonerTier string
	// {1, 2, 3, 4, 5} as integer
	SummonerDivision int

	NumDeaths     int
	MinionsKilled int
	WardsPlaced   int
	WardsCleared  int
	ChampionId	  int
}

/**
 * ProcessedLeagueDocument
 */

type ProcessedLeagueDocument struct {
	Metadata 	DocumentMetadata

	SummonerId int `bson:"_id"`
	LastUpdate time.Time

	Current    ProcessedLeagueRank
	Historical []ProcessedLeagueRank
}

func (pld ProcessedLeagueDocument) GetId() int {
	return pld.SummonerId
}

func (pld ProcessedLeagueDocument) GetCreationTime() time.Time {
	return pld.Metadata.CreationTime
}

func (pld ProcessedLeagueDocument) GetType() DocType {
	return ProcessedLeagueDoc
}

type ProcessedLeagueRank struct {
	// The approximate time that the user was promoted to this rank.
	LastKnown time.Time

	Tier     string
	Division int
}

/**
 * ProcessedSummonerDocument
 */
