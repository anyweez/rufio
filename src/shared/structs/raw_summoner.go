package structs

import (
	"time"
)

type SummonerResponseWrapper struct {
	Metadata ResponseMetadata
	Response map[string]RawSummonerResponse
}

type RawSummonerResponse struct {
	SummonerId    int `json:"id",_bson:"_id"`
	Name          string
	ProfileIconId int
	SummonerLevel int
	RevisionDate  int
}

func NewSummonerResponse() SummonerResponseWrapper {
	srw := SummonerResponseWrapper{}
	srw.Metadata.RequestTime = time.Now()

	return srw
}
