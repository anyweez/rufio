package structs

import ()

type ProcessedSummoner struct {
	// Basic summoner information.
	SummonerId int `bson:"_id"`
	Name       string
	// Latest known ranking.
	CurrentTier     string
	CurrentDivision int
	// Games that the summoner has been involved in.
	CompleteGameIds   []int
	IncompleteGameIds []int
	// TODO: Probably more, depending on what I can get from the JSON response.
}
