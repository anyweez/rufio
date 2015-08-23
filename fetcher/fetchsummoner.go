package main

import (
	"encoding/json"
	"fmt"
	proto "github.com/luke-segars/rufio/proto"
	structs "github.com/luke-segars/rufio/shared/structs"
	"io/ioutil"
	"net/http"
)

const SUMMONER_API_URL = "https://na.api.pvp.net//api/lol/na/v1.4/summoner/%d?api_key=%s"

func init() {
	configs = append(configs, FetcherConfig{
		FetchType:   "summoner",
		FetchParser: ParseSummoner,
		TubeName:    "retrieve_summoner_info",

		DatabaseName:   "league",
		CollectionName: "raw_summoners",

		BuildUrl: MakeSummonerUrl,
	})
}

func ParseSummoner(response *http.Response) (structs.RawResponseWrapper, error) {
	defer response.Body.Close()

	sr := structs.NewSummonerResponse()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return sr, err
	}

	// Parse and store the response.
	err = json.Unmarshal(body, &sr.Response)
	if err != nil {
		return sr, err
	}

	return sr, nil
}

func MakeSummonerUrl(req proto.ProcessedJobRequest, apiKey string) string {
	return fmt.Sprintf(SUMMONER_API_URL, *req.TargetId, apiKey)
}
