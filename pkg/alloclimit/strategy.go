package alloclimit

import (
	"context"

	"github.com/dgraph-io/badger/v3"

	"github.com/Blinkuu/qms/pkg/alloclimit/local"
	"github.com/Blinkuu/qms/pkg/alloclimit/memory"
)

type StrategyConfig struct {
	Capacity int64 `mapstructure:"capacity"`
}

type Strategy interface {
	Alloc(ctx context.Context, tokens int64) (remainingTokens int64, ok bool, err error)
	Free(ctx context.Context, tokens int64) (remainingTokens int64, ok bool, err error)
}

type StrategyFactory interface {
	Strategy(id string, capacity int64) (Strategy, error)
}

type MemoryStrategyFactory struct{}

func NewMemoryStrategyFactory() *MemoryStrategyFactory {
	return &MemoryStrategyFactory{}
}

func (*MemoryStrategyFactory) Strategy(_ string, capacity int64) (Strategy, error) {
	return memory.NewCappedBucket(capacity), nil
}

type LocalStrategyFactory struct {
	db *badger.DB
}

func NewLocalStrategyFactory(db *badger.DB) *LocalStrategyFactory {
	return &LocalStrategyFactory{
		db: db,
	}
}

func (f *LocalStrategyFactory) Strategy(id string, capacity int64) (Strategy, error) {
	return local.NewCappedBucket(f.db, []byte(id), capacity)
}
