package alloc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/internal/core/ports"
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
	storage, err := newStorageFromConfig(cfg, logger, memberlist)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage from config: %w", err)
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

		storage = raftStorage
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
