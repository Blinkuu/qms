package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"go.uber.org/zap"

	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/ratelimit"
	"github.com/Blinkuu/qms/pkg/ratelimit/memory"
	"github.com/Blinkuu/qms/pkg/timeunit"
)

type strategyDefinition struct {
	Type string `mapstructure:"type"`

	// Rate type
	Algorithm      string `mapstructure:"algorithm,omitempty"`
	Unit           string `mapstructure:"unit,omitempty"`
	RequestPerUnit int64  `mapstructure:"requests_per_unit,omitempty"`
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
	strategy ratelimit.Strategy
}

type guardContainer map[string]*guard

type QuotaService struct {
	guardContainer   guardContainer
	guardContainerMu *sync.RWMutex
}

func NewQuotaService(clock clock.Clock, logger log.Logger, cfg QuotaServiceConfig) *QuotaService {
	guardContainer := make(guardContainer)

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := guardContainer[id]
		if found {
			logger.Panic("only single namespace-resource quota can be registered")
		}

		if quota.Strategy.Type != "rate" {
			logger.Panic("strategy type not supported", zap.String("type", quota.Strategy.Type))
		}

		if quota.Strategy.Algorithm != "token-bucket" {
			logger.Panic("algorithm type not supported", zap.String("algorithm", quota.Strategy.Type))
		}

		unit, err := timeunit.Parse(quota.Strategy.Unit)
		if err != nil {
			logger.Panic("failed to parse time unit", zap.Error(err))
		}

		guardContainer[id] = &guard{
			strategy: memory.NewTokenBucket(
				quota.Strategy.RequestPerUnit*int64(unit),
				quota.Strategy.RequestPerUnit*int64(unit),
				clock,
			),
		}
	}

	return &QuotaService{
		guardContainer:   guardContainer,
		guardContainerMu: &sync.RWMutex{},
	}
}

func (q *QuotaService) Allow(namespace string, resource string, tokens int64) (time.Duration, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	q.guardContainerMu.RLock()
	defer q.guardContainerMu.RUnlock()

	guard, found := q.guardContainer[id]
	if !found {
		return 0, fmt.Errorf("guard for %s not found", id)
	}

	waitTime, err := guard.strategy.Allow(tokens)
	if err != nil {
		return 0, fmt.Errorf("failed to allow: %w", err)
	}

	return waitTime, nil
}
