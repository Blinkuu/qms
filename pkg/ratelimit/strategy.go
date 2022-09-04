package ratelimit

import "time"

type Strategy interface {
	Allow(tokens int64) (waitTime time.Duration, err error)
}
