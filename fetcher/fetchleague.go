package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	proto "github.com/luke-segars/rufio/proto"
	structs "github.com/luke-segars/rufio/shared/structs"
)

// Format of the game request URL.
var LEAGUE_API_URL = "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/%d?api_key=%s"

func init() {
	configs = append(configs, FetcherConfig{
		FetchType: "league",
		FetchParser: ParseLeague,
		TubeName: "retrieve_recent_league",
	
		DatabaseName: "league",
		CollectionName: "raw_leagues",

		BuildUrl: MakeLeagueUrl,
	})
}

func ParseLeague(response *http.Response) (structs.RawResponseWrapper, error) {
	defer response.Body.Close()

	gr := structs.NewLeagueResponse()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return gr, err
	}

	json.Unmarshal(body, &gr.Response)
	if err != nil {
		return gr, err
	}

	return gr, nil
}

func MakeLeagueUrl(req proto.ProcessedJobRequest, apiKey string) string {
	return fmt.Sprintf(LEAGUE_API_URL, *req.TargetId, apiKey)
}