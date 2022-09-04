package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Blinkuu/qms/pkg/alloclimit"
	"github.com/Blinkuu/qms/pkg/log"
)

type allocationStrategyDefinition struct {
	Capacity int64 `mapstructure:"capacity,omitempty"`
}

type allocationQuotaDefinition struct {
	Namespace string                       `mapstructure:"namespace"`
	Resource  string                       `mapstructure:"resource"`
	Strategy  allocationStrategyDefinition `mapstructure:"strategy"`
}

type AllocationQuotaServiceConfig struct {
	Backend string                      `mapstructure:"backend"`
	Quotas  []allocationQuotaDefinition `mapstructure:"quotas"`
}

type guard[T any] struct {
	strategy T
}

type allocationGuardContainer map[string]*guard[alloclimit.Strategy]

type AllocationQuotaService struct {
	logger           log.Logger
	guardContainer   allocationGuardContainer
	guardContainerMu *sync.RWMutex
}

func NewAllocationQuotaService(logger log.Logger, cfg AllocationQuotaServiceConfig) (*AllocationQuotaService, error) {
	guardContainer, err := allocationGuardContainerFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create guard container from config: %w", err)
	}

	return &AllocationQuotaService{
		logger:           logger,
		guardContainer:   guardContainer,
		guardContainerMu: &sync.RWMutex{},
	}, nil
}

func (q *AllocationQuotaService) Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	q.logger.Info("alloc called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, false, fmt.Errorf("guard for %s not found", id)
	}

	remainingTokens, ok, err := guard.strategy.Alloc(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, ok, nil
}

func (q *AllocationQuotaService) Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	q.logger.Info("free called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, false, fmt.Errorf("guard for %s not found", id)
	}

	remainingTokens, ok, err := guard.strategy.Free(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return remainingTokens, ok, nil
}

func allocationGuardContainerFromConfig(cfg AllocationQuotaServiceConfig) (allocationGuardContainer, error) {
	guardContainer := make(allocationGuardContainer)

	var strategyFactory alloclimit.StrategyFactory
	switch cfg.Backend {
	case "memory":
		strategyFactory = alloclimit.NewMemoryStrategyFactory()
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Backend)
	}

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := guardContainer[id]
		if found {
			return nil, errors.New("only a single namespace-resource pair can be registered")
		}

		strategy, err := strategyFactory.Strategy(quota.Strategy.Capacity)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ratelimit strategy: %w", err)
		}

		guardContainer[id] = &guard[alloclimit.Strategy]{
			strategy: strategy,
		}
	}

	return guardContainer, nil
}
