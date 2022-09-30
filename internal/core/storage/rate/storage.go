package rate

import (
	"context"
	"time"

	"github.com/Blinkuu/qms/internal/core/storage/rate/quota"
)

type Storage interface {
	Allow(ctx context.Context, namespace, resource string, tokens int64) (waitTime time.Duration, ok bool, err error)
	RegisterQuota(namespace, resource string, cfg quota.Config) error
}
