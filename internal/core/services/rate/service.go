package rate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/ratelimit"
)

const (
	ServiceName = "rate"
)

type Service struct {
	services.NamedService
	logger       log.Logger
	strategies   map[string]ratelimit.Strategy
	strategiesMu *sync.RWMutex
}

func NewService(cfg Config, clock clock.Clock, logger log.Logger) (*Service, error) {
	logger = logger.With("service", ServiceName)

	guardContainer, err := ratelimitStrategiesFromConfig(cfg, clock)
	if err != nil {
		return nil, fmt.Errorf("failed to create ratelimit strategies from config: %w", err)
	}

	s := &Service{
		NamedService: nil,
		logger:       logger,
		strategies:   guardContainer,
		strategiesMu: &sync.RWMutex{},
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	hostname, _ := os.Hostname()
	s.logger.Info("allow called", "namespace", namespace, "resource", resource, "tokens", tokens, "hostname", hostname)

	id := strings.Join([]string{namespace, resource}, "_")

	s.strategiesMu.RLock()
	defer s.strategiesMu.RUnlock()

	strategy, found := s.strategies[id]
	if !found {
		return 0, false, fmt.Errorf("strategy for %s not found", id)
	}

	waitTime, ok, err := strategy.Allow(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to allow: %w", err)
	}

	return waitTime, ok, nil
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting rate service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running rate service")

	<-ctx.Done()

	return nil
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping rate service")

	if err != nil {
		s.logger.Error("rate service returned error from running state", "err", err)
	}

	return nil
}

func ratelimitStrategiesFromConfig(cfg Config, clock clock.Clock) (map[string]ratelimit.Strategy, error) {
	strategies := make(map[string]ratelimit.Strategy, len(cfg.Quotas))

	var strategyFactory ratelimit.StrategyFactory
	switch cfg.Backend {
	case "memory":
		strategyFactory = ratelimit.NewMemoryStrategyFactory(clock)
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Backend)
	}

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := strategies[id]
		if found {
			return nil, errors.New("only a single strategy for a namespace-resource pair can be registered")
		}

		strategy, err := strategyFactory.Strategy(quota.Strategy.Algorithm, quota.Strategy.Unit, quota.Strategy.RequestPerUnit)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ratelimit strategy: %w", err)
		}

		strategies[id] = strategy
	}

	return strategies, nil
}
