package alloc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/internal/core/storage"
	"github.com/Blinkuu/qms/internal/core/storage/alloc"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/local"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/memory"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/raft"
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

func NewService(cfg Config, logger log.Logger, memberlist ports.MemberlistService) (*Service, error) {
	st, err := newStorageFromConfig(cfg, logger, memberlist)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage from config: %w", err)
	}

	s := &Service{
		NamedService: nil,
		logger:       logger,
		storage:      st,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) View(ctx context.Context, namespace, resource string) (int64, int64, int64, error) {
	allocated, capacity, version, err := s.storage.View(ctx, namespace, resource)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return 0, 0, 0, ErrNotFound
		default:
		}

		return 0, 0, 0, fmt.Errorf("failed to view: %w", err)
	}

	return allocated, capacity, version, nil
}

func (s *Service) Alloc(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	remainingTokens, currentVersion, ok, err := s.storage.Alloc(ctx, namespace, resource, tokens, version)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return 0, 0, false, ErrNotFound
		case errors.Is(err, storage.ErrInvalidVersion):
			return 0, 0, false, ErrInvalidVersion
		default:
		}

		return 0, 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, currentVersion, ok, nil
}

func (s *Service) Free(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	remainingTokens, currentVersion, ok, err := s.storage.Free(ctx, namespace, resource, tokens, version)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return 0, 0, false, ErrNotFound
		case errors.Is(err, storage.ErrInvalidVersion):
			return 0, 0, false, ErrInvalidVersion
		default:
		}

		return 0, 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, currentVersion, ok, nil
}

func (s *Service) Join(ctx context.Context, replicaID uint64, raftAddr string) (bool, error) {
	raftStorage, ok := s.storage.(*raft.Storage)
	if !ok {
		return false, errors.New("underlying storage is not a raft storage")
	}

	return raftStorage.AddRaftReplica(ctx, replicaID, raftAddr)
}

func (s *Service) Exit(ctx context.Context, replicaID uint64) error {
	raftStorage, ok := s.storage.(*raft.Storage)
	if !ok {
		return errors.New("underlying storage is not a raft storage")
	}

	return raftStorage.RemoveRaftReplica(ctx, replicaID)
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

	return s.storage.Shutdown(context.TODO())
}

func newStorageFromConfig(cfg Config, logger log.Logger, memberlist ports.MemberlistService) (alloc.Storage, error) {
	var st alloc.Storage
	switch cfg.Storage.Backend {
	case alloc.Memory:
		st = memory.NewStorage()
	case alloc.Local:
		var err error
		st, err = local.NewStorage(cfg.Storage.Local, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create new local storage: %w", err)
		}
	case alloc.Raft:
		raftStorage, err := raft.NewStorage(cfg.Storage.Raft, logger, memberlist)
		if err != nil {
			return nil, fmt.Errorf("failed to create new raft storage: %w", err)
		}

		go func() {
			if err := raftStorage.Run(context.Background()); err != nil {
				logger.Panic("failed to run raft storage", "err", err)
			}
		}()

		if err := raftStorage.AwaitHealthy(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to await healthy for raft storage: %w", err)
		}

		st = raftStorage
	default:
		return nil, fmt.Errorf("%s backend is not supported", cfg.Storage.Backend)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, quota := range cfg.Quotas {
		err := st.RegisterQuota(ctx, quota.Namespace, quota.Resource, quota.Strategy)
		if err != nil {
			logger.Warn("failed to register quota", "err", err)
		}
	}

	return st, nil
}
