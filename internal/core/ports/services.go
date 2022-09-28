package ports

import (
	"context"
	"time"

	"github.com/Blinkuu/qms/internal/core/domain"
)

type PingService interface {
	Ping(ctx context.Context) string
}

type MemberlistService interface {
	Members(ctx context.Context) ([]domain.Instance, error)
}

type KVService interface {
	Get(ctx context.Context, key string) (any, error)

	Delete(ctx context.Context, key string) error

	CAS(ctx context.Context, key string, f func(in any) (out any, retry bool, err error)) error

	WatchKey(ctx context.Context, key string, f func(interface{}) bool)
}

type RateService interface {
	Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error)
}

type AllocService interface {
	Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error)
	Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error)
}

type ProxyService interface {
	RateService
	AllocService
}
