package services

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/pkg/alloclimit"
	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/ratelimit"
)

type strategyDefinition struct {
	Type string `mapstructure:"type"`

	// Rate type
	Algorithm      string `mapstructure:"algorithm,omitempty"`
	Unit           string `mapstructure:"unit,omitempty"`
	RequestPerUnit int64  `mapstructure:"requests_per_unit,omitempty"`

	// Allocation type
	Capacity int64 `mapstructure:"capacity,omitempty"`
}

type quotaDefinition struct {
	Namespace string             `mapstructure:"namespace"`
	Resource  string             `mapstructure:"resource"`
	Strategy  strategyDefinition `mapstructure:"strategy"`
}

type QuotaServiceConfig struct {
	Quotas []quotaDefinition `mapstructure:"quotas"`
}

type guard struct {
	ratelimitStrategy   ratelimit.Strategy
	alloclimitStrategy  alloclimit.Strategy
	isRateLimitStrategy bool
}

type guardContainer map[string]*guard

type QuotaService struct {
	logger           log.Logger
	guardContainer   guardContainer
	guardContainerMu *sync.RWMutex
}

func NewQuotaService(clock clock.Clock, logger log.Logger, cfg QuotaServiceConfig) (*QuotaService, error) {
	guardContainer, err := guardContainerFromConfig(clock, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create guard container from config: %w", err)
	}

	return &QuotaService{
		logger:           logger,
		guardContainer:   guardContainer,
		guardContainerMu: &sync.RWMutex{},
	}, nil
}

func (q *QuotaService) Allow(namespace string, resource string, tokens int64) (time.Duration, error) {
	q.logger.Info("allow called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, fmt.Errorf("guard for %s not found", id)
	}

	if !guard.isRateLimitStrategy {
		return 0, fmt.Errorf("rate limit strategy is not registered for %s", id)
	}

	waitTime, err := guard.ratelimitStrategy.Allow(tokens)
	if err != nil {
		return 0, fmt.Errorf("failed to allow: %w", err)
	}

	return waitTime, nil
}

func (q *QuotaService) Alloc(namespace string, resource string, tokens int64) (int64, bool, error) {
	q.logger.Info("alloc called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, false, fmt.Errorf("guard for %s not found", id)
	}

	if guard.isRateLimitStrategy {
		return 0, false, fmt.Errorf("alloc strategy is not registered for %s", id)
	}

	remainingTokens, ok, err := guard.alloclimitStrategy.Alloc(tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, ok, nil
}

func (q *QuotaService) Free(namespace string, resource string, tokens int64) (int64, bool, error) {
	q.logger.Info("free called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, false, fmt.Errorf("guard for %s not found", id)
	}

	if guard.isRateLimitStrategy {
		return 0, false, fmt.Errorf("alloc strategy is not registered for %s", id)
	}

	remainingTokens, ok, err := guard.alloclimitStrategy.Free(tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return remainingTokens, ok, nil
}

func guardContainerFromConfig(clock clock.Clock, cfg QuotaServiceConfig) (guardContainer, error) {
	guardContainer := make(guardContainer)

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := guardContainer[id]
		if found {
			return nil, errors.New("only a single namespace-resource pair can be registered")
		}

		switch quota.Strategy.Type {
		case "rate":
			strategy, err := ratelimit.
				NewStrategyFactory(clock).
				Strategy(
					quota.Strategy.Algorithm,
					quota.Strategy.Unit,
					quota.Strategy.RequestPerUnit,
				)
			if err != nil {
				return nil, fmt.Errorf("failed to create new ratelimit strategy: %w", err)
			}

			guardContainer[id] = &guard{
				ratelimitStrategy:   strategy,
				isRateLimitStrategy: true,
			}
		case "allocation":
			strategy, err := alloclimit.
				NewStrategyFactory().
				Strategy(quota.Strategy.Capacity)
			if err != nil {
				return nil, fmt.Errorf("failed to create new alloclimit strategy: %w", err)
			}

			guardContainer[id] = &guard{
				alloclimitStrategy:  strategy,
				isRateLimitStrategy: false,
			}
		default:
			return nil, fmt.Errorf("strategy type %s is not supported", quota.Strategy.Type)
		}
	}

	return guardContainer, nil
}
