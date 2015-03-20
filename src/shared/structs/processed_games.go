package structs

type ProcessedGame struct {
	GameId        int `bson:"_id"`
	GameTimestamp int64
	// String representation of the above timestamp in "YYYY-MM-DD" format.
	GameDate string
	Stats    []ProcessedPlayerStats
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
}
