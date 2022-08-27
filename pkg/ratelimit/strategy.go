package ratelimit

import "time"

type RateStrategy interface {
	Allow(weight int64) (time.Duration, error)
}
