package ports

import (
	"context"
	"time"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/domain"
)

type PingService interface {
	services.NamedService
	Ping(ctx context.Context) string
}

type MemberlistService interface {
	services.NamedService
	Members(ctx context.Context) ([]domain.Instance, error)
}

type RateService interface {
	services.NamedService
	Allow(ctx context.Context, namespace, resource string, tokens int64) (waitTime time.Duration, ok bool, err error)
}

type AllocService interface {
	services.NamedService
	Alloc(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error)
	Free(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error)
}

type ProxyService interface {
	services.NamedService
	RateService
	AllocService
}

type RaftService interface {
	Join(ctx context.Context, replicaID uint64, raftAddr string) (alreadyMember bool, err error)
	Exit(ctx context.Context, replicaID uint64) error
}
