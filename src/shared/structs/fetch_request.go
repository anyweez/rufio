package structs

import (
	proto "proto"
)

type FetchRequest struct {
	Job proto.ProcessedJobRequest
	Url string
}
