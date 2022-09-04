package ports

import (
	"context"
	"time"
)

type PingService interface {
	Ping(ctx context.Context) string
}

type QuotaService interface {
	Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, error)
	Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error)
	Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error)
}
