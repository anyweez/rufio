package processed

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/luke-segars/loldata/shared/structs"
)

func (api *ProcessedApi) StoreLeague(league structs.ProcessedLeague) error {
	c := api.Session.DB("league").C("processed_leagues")
	_, err := c.Upsert(bson.M{"_id": league.SummonerId}, league)

	return err
}

func (api *ProcessedApi) StoreGame(game structs.ProcessedGame) error {
	c := api.Session.DB("league").C("processed_games")
	_, err := c.Upsert(bson.M{"_id": game.GameId}, game)

	return err
}