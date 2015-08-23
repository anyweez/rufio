package structs

import (
	"time"
)

type RawResponseWrapper interface {
	When() time.Time
}
