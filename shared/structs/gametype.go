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
 * This function maps Riot's game types to loldata game types (fewer, represented with ints
 *	instead of string).
 *
 * Full list of Riot game types at: https://developer.riotgames.com/docs/game-constants.
 */
func GetGameType(raw GameResponseGame) ProcessedGameType {
	switch raw.SubType {
	case "NORMAL_5x5_BLIND":
		return NORMAL_5X5_SR
	case "NORMAL_5x5_DRAFT":
		return NORMAL_5X5_SR
	case "GROUP_FINDER_5x5": // Team Builder
		return NORMAL_5X5_SR
	case "RANKED_SOLO_5x5":
		return RANKED_SOLO_5X5_SR
	case "RANKED_TEAM_5x5":
		return RANKED_TEAM_5X5_SR
	case "RANKED_PREMADE_5x5":
		return RANKED_TEAM_5X5_SR

	case "NORMAL_3x3":
		return NORMAL_3X3
	case "RANKED_PREMADE_3x3":
		return RANKED_3X3
	case "RANKED_TEAM_3x3":
		return RANKED_3X3

	case "ODIN_5x5_BLIND":
		return NORMAL_5X5_D
	case "ODIN_5x5_DRAFT":
		return NORMAL_5X5_D

	default:
		return UNKNOWN_GAME
	}
}

/**
 * Function to centralize the definition of "summoner's rift games."
 */
func IsSummonersRiftGame(game ProcessedGameDocument) bool {
	return game.GameType == RANKED_SOLO_5X5_SR ||
		game.GameType == RANKED_TEAM_5X5_SR ||
		game.GameType == NORMAL_5X5_SR
}