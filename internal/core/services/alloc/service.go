package alloc

import (
	"context"
	"fmt"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/storage/alloc"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/local"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/memory"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "alloc"
)

type Service struct {
	services.NamedService
	logger  log.Logger
	storage alloc.Storage
}

func NewService(cfg Config, logger log.Logger) (*Service, error) {
	logger = logger.With("service", ServiceName)

	storage, err := newStorageFromConfig(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to alloclimit storages from config: %w", err)
	}

	s := &Service{
		NamedService: nil,
		logger:       logger,
		storage:      storage,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.logger.Info("alloc called", "namespace", namespace, "resource", resource, "tokens", tokens)

	return s.storage.Alloc(ctx, namespace, resource, tokens)
}

func (s *Service) Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.logger.Info("free called", "namespace", namespace, "resource", resource, "tokens", tokens)

	return s.storage.Free(ctx, namespace, resource, tokens)
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

func newStorageFromConfig(cfg Config, logger log.Logger) (alloc.Storage, error) {
	var storage alloc.Storage
	switch cfg.Storage.Backend {
	case alloc.Memory:
		storage = memory.NewStorage()
	case alloc.Local:
		var err error
		storage, err = local.NewStorage(cfg.Storage.Local, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create new local storage: %w", err)
		}
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
