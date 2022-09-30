package alloc

import (
	"context"

	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
)

type Storage interface {
	Alloc(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error)
	Free(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error)
	RegisterQuota(namespace, resource string, cfg quota.Config) error
}
