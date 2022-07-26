package alloc

import (
	"context"

	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
)

type Storage interface {
	View(ctx context.Context, namespace, resource string) (allocated, capacity, version int64, err error)
	Alloc(ctx context.Context, namespace, resource string, tokens, version int64) (remainingTokens, currentVersion int64, ok bool, err error)
	Free(ctx context.Context, namespace, resource string, tokens, version int64) (remainingTokens, currentVersion int64, ok bool, err error)
	RegisterQuota(ctx context.Context, namespace, resource string, cfg quota.Config) error
	Shutdown(ctx context.Context) error
}
