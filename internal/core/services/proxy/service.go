package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/dskit/services"
	"github.com/serialx/hashring"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/internal/core/ports"
	"github.com/Blinkuu/qms/pkg/cloud"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "proxy"
)

type Service struct {
	services.NamedService
	cfg              Config
	logger           log.Logger
	discoverer       cloud.Discoverer
	memberlistClient ports.MemberlistServiceClient

	rateClient     ports.RateServiceClient
	rateMembers    []domain.Instance
	rateHashRing   *hashring.HashRing
	rateHashRingMu *sync.RWMutex

	allocClient     ports.AllocServiceClient
	allocMembers    []domain.Instance
	allocHashRing   *hashring.HashRing
	allocHashRingMu *sync.RWMutex
}

func NewService(cfg Config, logger log.Logger, discoverer cloud.Discoverer, memberlistClient ports.MemberlistServiceClient, rateClient ports.RateServiceClient, allocClient ports.AllocServiceClient) (*Service, error) {
	s := &Service{
		NamedService:     nil,
		cfg:              cfg,
		logger:           logger,
		discoverer:       discoverer,
		memberlistClient: memberlistClient,
		rateClient:       rateClient,
		rateMembers:      nil,
		rateHashRing:     hashring.New(nil),
		rateHashRingMu:   &sync.RWMutex{},
		allocClient:      allocClient,
		allocMembers:     nil,
		allocHashRing:    hashring.New(nil),
		allocHashRingMu:  &sync.RWMutex{},
	}

	// TODO: Hash ring should be implemented using static node IDs
	// TODO: First, implement proxy service using only memberlistClient (simple hashing?). Then, implement a ring.Service
	//		 that uses a kv.Service to build a hash ring and swap

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	s.rateHashRingMu.RLock()
	defer s.rateHashRingMu.RUnlock()

	id := strings.Join([]string{namespace, resource}, "_")
	addr, ok := s.rateHashRing.GetNode(id)
	if !ok {
		return 0, false, fmt.Errorf("failed to get address from ring: id=%s", id)
	}

	return s.rateClient.Allow(ctx, []string{addr}, namespace, resource, tokens)
}

func (s *Service) Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.allocHashRingMu.RLock()
	defer s.allocHashRingMu.RUnlock()

	id := strings.Join([]string{namespace, resource}, "_")
	addr, ok := s.allocHashRing.GetNode(id)
	if !ok {
		return 0, false, fmt.Errorf("failed to get address from ring: id=%s", id)
	}

	return s.allocClient.Alloc(ctx, []string{addr}, namespace, resource, tokens)
}

func (s *Service) Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	s.allocHashRingMu.RLock()
	defer s.allocHashRingMu.RUnlock()

	id := strings.Join([]string{namespace, resource}, "_")
	addr, ok := s.allocHashRing.GetNode(id)
	if !ok {
		return 0, false, fmt.Errorf("failed to get address from ring: id=%s", id)
	}

	return s.allocClient.Free(ctx, []string{addr}, namespace, resource, tokens)
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting proxy service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running proxy service")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.updateRings()
		}
	}
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping proxy service")

	if err != nil {
		s.logger.Error("proxy service returned error from running state", "err", err)
	}

	return nil
}

func (s *Service) updateRings() {
	members, err := s.fetchMembers(s.cfg.RateAddresses)
	if err != nil {
		s.logger.Warn("failed to get rate members", "err", err)
	} else {
		s.updateRateHashRing(members)
	}

	members, err = s.fetchMembers(s.cfg.AllocAddresses)
	if err != nil {
		s.logger.Warn("failed to get alloc members", "err", err)
	} else {
		s.updateAllocHashRing(members)
	}
}

func (s *Service) fetchMembers(discoverAddrs []string) ([]domain.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	instances, err := s.discoverer.Discover(ctx, discoverAddrs)
	if err != nil {
		return nil, fmt.Errorf("failed to discover: %w", err)
	}

	addrs := make([]string, 0, len(instances))
	for _, instance := range instances {
		addrs = append(addrs, net.JoinHostPort(instance.Host, strconv.Itoa(instance.HTTPPort)))
	}

	members, err := s.memberlistClient.Members(ctx, addrs)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return members, nil
}

func (s *Service) updateRateHashRing(newMembers []domain.Instance) {
	s.rateHashRingMu.Lock()
	defer s.rateHashRingMu.Unlock()

	oldMembers := s.rateMembers
	diff := diffMembers(oldMembers, newMembers)

	hashRing := s.rateHashRing
	for _, added := range diff.added {
		hashRing = hashRing.AddWeightedNode(fmt.Sprintf("%s:%d", added.Host, added.HTTPPort), 1)
	}

	for _, removed := range diff.removed {
		hashRing = hashRing.RemoveNode(fmt.Sprintf("%s:%d", removed.Host, removed.HTTPPort))
	}

	s.rateMembers = newMembers
	s.rateHashRing = hashRing
}

func (s *Service) updateAllocHashRing(newMembers []domain.Instance) {
	s.allocHashRingMu.Lock()
	defer s.allocHashRingMu.Unlock()

	oldMembers := s.allocMembers
	diff := diffMembers(oldMembers, newMembers)

	hashRing := s.allocHashRing
	for _, added := range diff.added {
		hashRing = hashRing.AddWeightedNode(fmt.Sprintf("%s:%d", added.Host, added.HTTPPort), 1)
	}

	for _, removed := range diff.removed {
		hashRing = hashRing.RemoveNode(fmt.Sprintf("%s:%d", removed.Host, removed.HTTPPort))
	}

	s.allocMembers = newMembers
	s.allocHashRing = hashRing
}

type membersDiff struct {
	added   []domain.Instance
	removed []domain.Instance
}

func diffMembers(oldMembers, newMembers []domain.Instance) membersDiff {
	newMembersSet := make(map[domain.Instance]struct{}, len(newMembers))
	for _, newMember := range newMembers {
		newMembersSet[newMember] = struct{}{}
	}

	oldMembersSet := make(map[domain.Instance]struct{}, len(oldMembers))
	for _, oldMember := range oldMembers {
		oldMembersSet[oldMember] = struct{}{}
	}

	diff := membersDiff{}
	for _, oldMember := range oldMembers {
		if _, found := newMembersSet[oldMember]; !found {
			diff.removed = append(diff.removed, oldMember)
		}
	}

	for _, newMember := range newMembers {
		if _, found := oldMembersSet[newMember]; !found {
			diff.added = append(diff.added, newMember)
		}
	}

	return diff
}
