package processed

import (
	mgo "gopkg.in/mgo.v2"
	"time"
)

type ProcessedApi struct {
	Session *mgo.Session
}

func NewProcessedApi(connection string) (ProcessedApi, error) {
	api := ProcessedApi{}
	session, err := mgo.Dial(connection)

	if err != nil {
		return api, err
	}

	session.SetSocketTimeout(10 * time.Hour)
	api.Session = session
	return api, nil
}
