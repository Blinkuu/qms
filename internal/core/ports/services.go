package ports

import "time"

type PingService interface {
	Ping() string
}

type QuotaService interface {
	Allow(namespace string, resource string, weight int64) (time.Duration, error)
}
