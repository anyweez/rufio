package main

import (
	"encoding/json"
	"fmt"
	proto "github.com/luke-segars/rufio/proto"
	structs "github.com/luke-segars/rufio/shared/structs"
	"io/ioutil"
	"net/http"
)

// Format of the game request URL.
var GAME_API_URL = "https://na.api.pvp.net/api/lol/na/v1.3/game/by-summoner/%d/recent?api_key=%s"

func init() {
	configs = append(configs, FetcherConfig{
		FetchType:   "game",
		FetchParser: ParseGame,
		TubeName:    "retrieve_recent_games",

		DatabaseName:   "league",
		CollectionName: "raw_games",

		BuildUrl: MakeGameUrl,
	})
}

/**
 * Extract the game information from the HTTP response. This function will only be invoked
 * on a response that returned with a 200 status code (success), so no need to account for
 * other statuses.
 */
func ParseGame(response *http.Response) (structs.RawResponseWrapper, error) {
	defer response.Body.Close()

	game := structs.NewGameResponse()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return game, err
	}

	err = json.Unmarshal(body, &game.Response)
	if err != nil {
		return game, err
	}

	return game, nil
}

func MakeGameUrl(req proto.ProcessedJobRequest, apiKey string) string {
	return fmt.Sprintf(GAME_API_URL, *req.TargetId, apiKey)
}
