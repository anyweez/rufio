package main

import (
	"flag"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

var MONGO_CONNECTION = flag.String("mongodb", "localhost", "The address of the MongoDB instance to pull from")
var LOOKBACK_DAYS = flag.Int("lookback", 14, "The number of days to look back.")

const DAY = 24 * time.Hour

func main() {
	flag.Parse()

	session, err := mgo.Dial(*MONGO_CONNECTION)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Connected!")

	collection := session.DB("league").C("raw_games")

	rg_data := make([]int, *LOOKBACK_DAYS)
	for i := 0; i < *LOOKBACK_DAYS; i++ {
		// Start at the back of the time range and work up towards today.
		num_hours := time.Duration(-24 * (*LOOKBACK_DAYS - i))
		begin := time.Now().Add(time.Hour * num_hours)
		end := begin.Add(DAY)

		rg_data[i], _ = collection.Find(bson.M{
			"metadata.requesttime": bson.M{
				"$gt": begin.Truncate(DAY),
				"$lt": end.Truncate(DAY),
			},
		}).Count()

		fmt.Println(fmt.Sprintf("%s: %d", begin.Truncate(DAY), rg_data[i]))
	}

	fmt.Println(rg_data)
}
