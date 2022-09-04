package ports

import "time"

type PingService interface {
	Ping() string
}

type QuotaService interface {
	Allow(namespace string, resource string, tokens int64) (time.Duration, error)
}
