package shared

import (
	"github.com/quipo/statsd"
)

var StatsLogger *statsd.StatsdClient

func init() {
	StatsLogger = statsd.NewStatsdClient("statsd.service.fy:8125", "league.")
	StatsLogger.CreateSocket()
}
