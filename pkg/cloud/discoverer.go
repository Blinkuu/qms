package cloud

import (
	"context"
)

type Discoverer interface {
	Discover(ctx context.Context, serviceNames []string) ([]string, error)
}
