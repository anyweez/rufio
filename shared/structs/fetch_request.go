package structs

import (
	proto "github.com/luke-segars/rufio/proto"
)

type FetchRequest struct {
	Job   proto.ProcessedJobRequest
	Queue string // the name of the queue that the event came from
	Url   string
}
