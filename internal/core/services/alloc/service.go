package alloc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/pkg/alloclimit"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "alloc"
)

type Service struct {
	services.NamedService
	logger       log.Logger
	strategies   map[string]alloclimit.Strategy
	strategiesMu *sync.RWMutex
}

func NewService(cfg Config, logger log.Logger) (*Service, error) {
	logger = logger.With("service", ServiceName)

	strategies, err := alloclimitStrategiesFromConfig(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to alloclimit strategies from config: %w", err)
	}

	s := &Service{
		NamedService: nil,
		logger:       logger,
		strategies:   strategies,
		strategiesMu: &sync.RWMutex{},
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.logger.Info("alloc called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	s.strategiesMu.RLock()
	defer s.strategiesMu.RUnlock()

	strategy, found := s.strategies[id]
	if !found {
		return 0, false, fmt.Errorf("strategy for %s not found", id)
	}

	remainingTokens, ok, err := strategy.Alloc(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, ok, nil
}

func (s *Service) Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.logger.Info("free called", "namespace", namespace, "resource", resource, "tokens", tokens)

	id := strings.Join([]string{namespace, resource}, "_")

	s.strategiesMu.RLock()
	defer s.strategiesMu.RUnlock()

	strategy, found := s.strategies[id]
	if !found {
		return 0, false, fmt.Errorf("strategy for %s not found", id)
	}

	remainingTokens, ok, err := strategy.Free(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return remainingTokens, ok, nil
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting alloc service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running alloc service")

	<-ctx.Done()

	return nil
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping alloc service")

	if err != nil {
		s.logger.Error("alloc service returned error from running state", "err", err)
	}

	return nil
}

func alloclimitStrategiesFromConfig(cfg Config, logger log.Logger) (map[string]alloclimit.Strategy, error) {
	strategies := make(map[string]alloclimit.Strategy, len(cfg.Quotas))

	var strategyFactory alloclimit.StrategyFactory
	switch cfg.Backend {
	case "memory":
		strategyFactory = alloclimit.NewMemoryStrategyFactory()
	case "local":
		opts := badger.DefaultOptions(cfg.Local.Dir)
		opts.Logger = badgerLogger{logger: logger}
		db, err := badger.Open(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to open badger: %w", err)
		}
		strategyFactory = alloclimit.NewLocalStrategyFactory(db)
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Backend)
	}

	for _, quota := range cfg.Quotas {
		id := strings.Join([]string{quota.Namespace, quota.Resource}, "_")

		_, found := strategies[id]
		if found {
			return nil, errors.New("only a single namespace-resource pair can be registered")
		}

		strategy, err := strategyFactory.Strategy(id, quota.Strategy.Capacity)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ratelimit strategy: %w", err)
		}

		strategies[id] = strategy
	}

	return strategies, nil
}

type badgerLogger struct {
	logger log.Logger
}

func (b badgerLogger) Debugf(template string, args ...any) {
	b.logger.Debugf(template, args)
}

func (b badgerLogger) Infof(template string, args ...any) {
	b.logger.Infof(template, args)
}

func (b badgerLogger) Warningf(template string, args ...any) {
	b.logger.Warnf(template, args)
}

func (b badgerLogger) Errorf(template string, args ...any) {
	b.logger.Errorf(template, args)
}
