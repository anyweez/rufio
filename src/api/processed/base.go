package processed

import (
	mgo "gopkg.in/mgo.v2"
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

	api.Session = session
	return api, nil
}
