package rate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/storage"
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
	storage, err := storageFromConfig(cfg, clock, logger)
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
	waitTime, ok, err := s.storage.Allow(ctx, namespace, resource, tokens)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return 0, false, ErrNotFound
		default:
		}

		return 0, false, fmt.Errorf("failed to view: %w", err)
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

	return s.storage.Shutdown(context.TODO())
}

func storageFromConfig(cfg Config, clock clock.Clock, logger log.Logger) (rate.Storage, error) {
	var storage rate.Storage
	switch cfg.Storage.Backend {
	case alloc.Memory:
		storage = memory.NewStorage(clock)
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Storage.Backend)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, quota := range cfg.Quotas {
		err := storage.RegisterQuota(ctx, quota.Namespace, quota.Resource, quota.Strategy)
		if err != nil {
			logger.Warn("failed to register quota", "err", err)
		}
	}

	return storage, nil
}
