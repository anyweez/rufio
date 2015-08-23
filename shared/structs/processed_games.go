package structs

type ProcessedGameType int

/**
 * These game types are a subset of all availalbe game types that Riot provides:
 * 		https://developer.riotgames.com/docs/game-constants
 * Game types that aren't represented here are labeled as "unknown."
 */
const (
	// Unknown
	UNKNOWN_GAME ProcessedGameType = iota

	// Summoner's Rift game types
	RANKED_SOLO_5X5_SR ProcessedGameType = iota
	RANKED_TEAM_5X5_SR ProcessedGameType = iota
	NORMAL_5X5_SR      ProcessedGameType = iota

	// Dominion game types
	NORMAL_5X5_D ProcessedGameType = iota

	// Twisted Treeline game types
	RANKED_3X3 ProcessedGameType = iota
	NORMAL_3X3 ProcessedGameType = iota
)

/**
 * Function to centralize the definition of "summoner's rift games."
 */
func IsSummonersRiftGame(game ProcessedGame) bool {
	return game.GameType == RANKED_SOLO_5X5_SR ||
		game.GameType == RANKED_TEAM_5X5_SR ||
		game.GameType == NORMAL_5X5_SR
}

type ProcessedGame struct {
	GameId        int `bson:"_id"`
	GameTimestamp int64
	// String representation of the above timestamp in "YYYY-MM-DD" format.
	GameDate string
	// Game type (see above)
	GameType ProcessedGameType

	Stats []ProcessedPlayerStats
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
	ChampionId    int
}
