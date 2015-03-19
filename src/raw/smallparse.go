package main

import (
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	"io/ioutil"
	"net/http"
	formats "raw/raw_leagues"
)

const GAME_URL = "https://na.api.pvp.net/api/lol/na/v1.3/game/by-summoner/36069121/recent?api_key=c34faa38-a0d4-4022-933e-c8cd81bf74fe"
const LEAGUE_URL = "https://na.api.pvp.net/api/lol/na/v2.5/league/by-summoner/36652890?api_key=c34faa38-a0d4-4022-933e-c8cd81bf74fe"

/*
type Tiers []TierRecord

type TierRecord struct {
	Name  string
	Tier  string
	Queue string
}
*/
func main() {
	resp, err := http.Get(LEAGUE_URL)
	if err != nil {
		fmt.Println("fetch error: " + err.Error())
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	// response := formats.NewLeagueResponse()

	// fmt.Println(string(body))
	response := formats.NewLeagueResponse()
	// response := make(map[string]Tiers)
	//	response := make(map[string]formats.LeagueResponseTier)
	json.Unmarshal(body, &response.Response)

	pretty.Println(response)
}
