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
	Allow(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (time.Duration, bool, error)
}

type AllocServiceClient interface {
	Alloc(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (int64, bool, error)
	Free(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (int64, bool, error)
}
