package shared

import (
	"github.com/luke-segars/rufio/shared/structs"
)

/**
 * This function maps Riot's game types to loldata game types (fewer, represented with ints
 *	instead of string).
 *
 * Full list of Riot game types at: https://developer.riotgames.com/docs/game-constants.
 */
func GetGameType(raw structs.GameResponseGame) structs.ProcessedGameType {
	switch raw.SubType {
	case "NORMAL_5x5_BLIND":
		return structs.NORMAL_5X5_SR
	case "NORMAL_5x5_DRAFT":
		return structs.NORMAL_5X5_SR
	case "GROUP_FINDER_5x5": // Team Builder
		return structs.NORMAL_5X5_SR
	case "RANKED_SOLO_5x5":
		return structs.RANKED_SOLO_5X5_SR
	case "RANKED_TEAM_5x5":
		return structs.RANKED_TEAM_5X5_SR
	case "RANKED_PREMADE_5x5":
		return structs.RANKED_TEAM_5X5_SR

	case "NORMAL_3x3":
		return structs.NORMAL_3X3
	case "RANKED_PREMADE_3x3":
		return structs.RANKED_3X3
	case "RANKED_TEAM_3x3":
		return structs.RANKED_3X3

	case "ODIN_5x5_BLIND":
		return structs.NORMAL_5X5_D
	case "ODIN_5x5_DRAFT":
		return structs.NORMAL_5X5_D

	default:
		return structs.UNKNOWN_GAME
	}
}
