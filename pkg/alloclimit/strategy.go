package alloclimit

import (
	"context"

	"github.com/Blinkuu/qms/pkg/alloclimit/memory"
)

type Strategy interface {
	Alloc(ctx context.Context, tokens int64) (remainingTokens int64, ok bool, err error)
	Free(ctx context.Context, tokens int64) (remainingTokens int64, ok bool, err error)
}

type StrategyFactory struct{}

func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{}
}

func (StrategyFactory) Strategy(capacity int64) (Strategy, error) {
	return memory.NewCappedBucket(capacity), nil
}
