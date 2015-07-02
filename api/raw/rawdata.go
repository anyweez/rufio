package rawdata

import (
	"errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	structs "shared/structs"
	strconv "strconv"
	"time"
)

// TODO: engineer this whole mess better.

type RawApi struct {
	Session *mgo.Session
}

func NewRawApi(connection string) (RawApi, error) {
	api := RawApi{}
	session, err := mgo.Dial(connection)
	// Prevent timeout errors unless things are REALLY bad.
	session.SetSocketTimeout(10 * time.Minute)

	if err != nil {
		return api, err
	}

	api.Session = session
	return api, nil
}

/**
 * Fetch a set of raw_games that include information related to the specified
 * game. Note that each GameResponse will only include data for a single player
 * so it's better to check processed_game logs if you're looking for more complete
 * game views.
 */
func (r *RawApi) GetPartialGames(gameid int) ([]structs.GameResponse, error) {
	gr := make([]structs.GameResponse, 0)
	collection := r.Session.DB("league").C("raw_games")
	iter := collection.Find(bson.M{"response.games.gameid": gameid}).Iter()

	result := structs.GameResponseWrapper{}

	for iter.Next(&result) {
		gr = append(gr, result.Response)
	}

	if iter.Err() != nil {
		log.Println("Game retrieval ERROR: " + iter.Err().Error())
		return gr, iter.Err()
	}

	return gr, nil
}

/**
 * Get a list of all complete games (games with stats available) for the provided
 * summoner ID.
 */
func (r *RawApi) GetCompleteGamesBySummoner(summoner_id int) []structs.GameResponse {
	gr := make([]structs.GameResponse, 0)
	collection := r.Session.DB("league").C("raw_games")
	iter := collection.Find(bson.M{"response.summonerid": summoner_id}).Iter()

	result := structs.GameResponseWrapper{}

	for iter.Next(&result) {
		gr = append(gr, result.Response)
	}

	return gr
}

func (r *RawApi) GetIncompleteGameIdsBySummoner(summoner_id int) []int {
	gameIds := make([]int, 0)
	collection := r.Session.DB("league").C("raw_games")
	iter := collection.Find(bson.M{
		"response.games.fellowplayers.summonerid": summoner_id,
	}).Select(bson.M{
		"response.games.gameid":                   true,
		"response.games.fellowplayers.summonerid": true,
	}).Iter()

	result := structs.GameResponseWrapper{}

	// Iterate through all results and find game ID's that contain this player as a
	// fellow player.
	for iter.Next(&result) {
		for _, game := range result.Response.Games {
			for _, player := range game.FellowPlayers {
				if player.SummonerId == summoner_id {
					gameIds = append(gameIds, game.GameId)
				}
			}
		}
	}

	return gameIds
}

/**
 * Returns the most recent league result for the provided summoner from PROCESSED_LEAGUES.
 */
/*
func (r *RawApi) GetLatestLeague(summoner_id int, queue_type string) (structs.ProcessedLeagueRank, error) {
	collection := r.Session.DB("league").C("processed_leagues")
	iter := collection.Find(bson.M{
		"_id": summoner_id,
	}).Iter()

	// Check to make sure that at least one result came back. If so, iterate through all results to
	// find the most recent one. If not, return an error.
	result := structs.LeagueResponseWrapper{}
	success := iter.Next(&result)

	if !success {
		return structs.LeagueResponseTier{}, errors.New("No matches found for summoner " + strconv.Itoa(summoner_id))
	}

	if result.Current == nil {
		return nil, errors.New("Unknown league rank.")
	}

	return result.Current, nil
}
*/
func (r *RawApi) GetRawSummonerInfo(summoner_id int) (structs.RawSummonerResponse, error) {
	collection := r.Session.DB("league").C("raw_summoners")
	// TODO: Get the most recent record (order by metadata.requesttime)
	iter := collection.Find(bson.M{
		"response." + strconv.Itoa(summoner_id): bson.M{"$exists": true},
	}).Iter()

	result := structs.SummonerResponseWrapper{}
	success := iter.Next(&result)

	if !success {
		return structs.RawSummonerResponse{}, errors.New("No matches found for summoner " + strconv.Itoa(summoner_id))
	}

	return result.Response[strconv.Itoa(summoner_id)], nil
}
