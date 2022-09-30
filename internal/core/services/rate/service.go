package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/storage/alloc"
	"github.com/Blinkuu/qms/internal/core/storage/rate"
	"github.com/Blinkuu/qms/internal/core/storage/rate/memory"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "rate"
)

type Service struct {
	services.NamedService
	logger  log.Logger
	storage rate.Storage
}

func NewService(cfg Config, clock clock.Clock, logger log.Logger) (*Service, error) {
	logger = logger.With("service", ServiceName)

	storage, err := storageFromConfig(cfg, clock)
	if err != nil {
		return nil, fmt.Errorf("failed create storage from config: %w", err)
	}

	s := &Service{
		NamedService: nil,
		logger:       logger,
		storage:      storage,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	s.logger.Info("allow called", "namespace", namespace, "resource", resource, "tokens", tokens)

	return s.storage.Allow(ctx, namespace, resource, tokens)
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

func storageFromConfig(cfg Config, clock clock.Clock) (rate.Storage, error) {
	var storage rate.Storage
	switch cfg.Storage.Backend {
	case alloc.Memory:
		storage = memory.NewStorage(clock)
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Storage.Backend)
	}

	for _, quota := range cfg.Quotas {
		err := storage.RegisterQuota(quota.Namespace, quota.Resource, quota.Strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to register quota: %w", err)
		}
	}

	return storage, nil
}
