package structs

import (
	proto "github.com/luke-segars/loldata/proto"
)

type FetchRequest struct {
	Job   proto.ProcessedJobRequest
	Queue string // the name of the queue that the event came from
	Url   string
}
