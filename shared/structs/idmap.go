package structs

import (
	"log"
	"errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type IDMap struct {
	mapping		map[int64][]int64
}

func NewIDMap() IDMap {
	idm := IDMap{}
	idm.mapping = make(map[int64][]int64)

	return idm
}

func (idm IDMap) add(key int64, val int64) {
	idm.mapping[key] = append(idm.mapping[key], val)
}

func (idm IDMap) Get(id int64) []int64 {
	return idm.mapping[id]
}

func LoadMappings(mongoUrl string, mongoDb string, types []string) (map[string]IDMap, error) {
	idm := make(map[string]IDMap)

	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return idm, err
	}

	db := session.DB(mongoDb)

	for _, mapName := range types {
		log.Println("Loading " +  mapName + " mapping...")

		switch (mapName) {
			// Maps game ID's to raw game document ID's
			case "game2rawgame":
				idm["game2rawgame"] = NewIDMap()

				doc := RawGameDocument{}
				result := db.C("raw_games").Find(bson.M{}).Iter()

				for result.Next(&doc) {
					for _, game := range doc.Response.Games {
						idm["game2rawgame"].add(int64(game.GameId), doc.Metadata.DocumentId)
					}
				}

				break;
			// Maps game ID's to summoner ID's for those who played in them THAT WE HAVE DATA FOR.
			case "game2summoner":
				log.Println("Loading game2summoner mapping...")
				idm["game2summoner"] = NewIDMap()

				doc := RawGameDocument{}
				result := db.C("raw_games").Find(bson.M{}).Iter()

				for result.Next(&doc) {
					for _, game := range doc.Response.Games {
						idm["game2rawgame"].add(int64(game.GameId), int64(doc.Response.SummonerId))
					}
				}

				break;
			default:
				return idm, errors.New("Unsupported mapping request.")
		}
	}

	return idm, nil
}