package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/ratelimit"
)

type rateStrategyDefinition struct {
	Algorithm      string `mapstructure:"algorithm"`
	Unit           string `mapstructure:"unit"`
	RequestPerUnit int64  `mapstructure:"requests_per_unit"`
}

type rateQuotaDefinition struct {
	Namespace string                 `mapstructure:"namespace"`
	Resource  string                 `mapstructure:"resource"`
	Strategy  rateStrategyDefinition `mapstructure:"strategy"`
}

type RateQuotaServiceConfig struct {
	Backend string                `mapstructure:"backend"`
	Quotas  []rateQuotaDefinition `mapstructure:"quotas"`
}

type ratelimitGuardContainer map[string]*guard[ratelimit.Strategy]

type RateQuotaService struct {
	logger           log.Logger
	guardContainer   ratelimitGuardContainer
	guardContainerMu *sync.RWMutex
}

func NewRateQuotaService(logger log.Logger, clock clock.Clock, cfg RateQuotaServiceConfig) (*RateQuotaService, error) {
	guardContainer, err := ratelimitGuardContainerFromConfig(clock, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create guard container from config: %w", err)
	}

	return &RateQuotaService{
		logger:           logger,
		guardContainer:   guardContainer,
		guardContainerMu: &sync.RWMutex{},
	}, nil
}

func (q *RateQuotaService) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, error) {
	q.logger.Info("allow called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, fmt.Errorf("guard for %s not found", id)
	}

	waitTime, err := guard.strategy.Allow(ctx, tokens)
	if err != nil {
		return 0, fmt.Errorf("failed to allow: %w", err)
	}

	return waitTime, nil
}

func ratelimitGuardContainerFromConfig(clock clock.Clock, cfg RateQuotaServiceConfig) (ratelimitGuardContainer, error) {
	guardContainer := make(ratelimitGuardContainer)

	var strategyFactory ratelimit.StrategyFactory
	switch cfg.Backend {
	case "memory":
		strategyFactory = ratelimit.NewMemoryStrategyFactory(clock)
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Backend)
	}

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := guardContainer[id]
		if found {
			return nil, errors.New("only a single namespace-resource pair can be registered")
		}

		strategy, err := strategyFactory.Strategy(quota.Strategy.Algorithm, quota.Strategy.Unit, quota.Strategy.RequestPerUnit)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ratelimit strategy: %w", err)
		}

		guardContainer[id] = &guard[ratelimit.Strategy]{
			strategy: strategy,
		}
	}

	return guardContainer, nil
}
