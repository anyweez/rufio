package main

import (
	"time"
)

type RateLimiter struct {
}

func NewRateLimiter(gapSize time.Duration) chan bool {
	rate := make(chan bool)

	go tick(gapSize, rate)

	return rate
}

func tick(gapSize time.Duration, limiter chan bool) {
	for {
		time.Sleep(gapSize)
		limiter <- true
	}
}
