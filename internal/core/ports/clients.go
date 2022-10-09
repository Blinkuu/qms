package ports

import (
	"context"
	"time"

	"github.com/Blinkuu/qms/internal/core/domain"
)

type MemberlistServiceClient interface {
	Members(ctx context.Context, addrs []string) ([]domain.Instance, error)
}

type RateServiceClient interface {
	Allow(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (waitTime time.Duration, ok bool, err error)
}

type AllocServiceClient interface {
	View(ctx context.Context, addrs []string, namespace, resource string) (allocated, capacity, version int64, err error)
	Alloc(ctx context.Context, addrs []string, namespace, resource string, tokens, version int64) (remainingTokens, currentVersion int64, ok bool, err error)
	Free(ctx context.Context, addrs []string, namespace, resource string, tokens, version int64) (remainingTokens, currentVersion int64, ok bool, err error)
}
