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

	rateClient   ports.RateServiceClient
	rateMembers  []domain.Instance
	rateHashRing *hashring.HashRing
	rateMu       *sync.RWMutex

	allocClient   ports.AllocServiceClient
	allocMembers  []domain.Instance
	allocHashRing *hashring.HashRing
	allocMu       *sync.RWMutex
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
		rateMu:           &sync.RWMutex{},
		allocClient:      allocClient,
		allocMembers:     nil,
		allocHashRing:    hashring.New(nil),
		allocMu:          &sync.RWMutex{},
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	s.rateMu.RLock()
	defer s.rateMu.RUnlock()

	addrs, err := s.hashRingLocked(namespace, resource)
	if err != nil {
		return 0, false, fmt.Errorf("failed to pick addresses from hash ring: %w", err)
	}

	return s.rateClient.Allow(ctx, addrs, namespace, resource, tokens)
}

func (s *Service) View(ctx context.Context, namespace, resource string) (int64, int64, int64, error) {
	s.allocMu.RLock()
	defer s.allocMu.RUnlock()

	var addrs []string
	switch s.cfg.AllocLBStrategy {
	case HashRingLBStrategy:
		a, err := s.hashRingLocked(namespace, resource)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to pick addresses from hash ring: %w", err)
		}

		addrs = a
	case RoundRobinLBStrategy:
		addrs = s.roundRobinLocked()
	default:
		return 0, 0, 0, fmt.Errorf("%s is not a supported alloc_lb_strategy", s.cfg.AllocLBStrategy)
	}

	return s.allocClient.View(ctx, addrs, namespace, resource)
}

func (s *Service) Alloc(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	s.allocMu.RLock()
	defer s.allocMu.RUnlock()

	var addrs []string
	switch s.cfg.AllocLBStrategy {
	case HashRingLBStrategy:
		a, err := s.hashRingLocked(namespace, resource)
		if err != nil {
			return 0, 0, false, fmt.Errorf("failed to pick addresses from hash ring: %w", err)
		}

		addrs = a
	case RoundRobinLBStrategy:
		addrs = s.roundRobinLocked()
	default:
		return 0, 0, false, fmt.Errorf("%s is not a supported alloc_lb_strategy", s.cfg.AllocLBStrategy)
	}

	return s.allocClient.Alloc(ctx, addrs, namespace, resource, tokens, version)
}

func (s *Service) Free(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	s.allocMu.RLock()
	defer s.allocMu.RUnlock()

	var addrs []string
	switch s.cfg.AllocLBStrategy {
	case HashRingLBStrategy:
		a, err := s.hashRingLocked(namespace, resource)
		if err != nil {
			return 0, 0, false, fmt.Errorf("failed to pick addresses from hash ring: %w", err)
		}

		addrs = a
	case RoundRobinLBStrategy:
		addrs = s.roundRobinLocked()
	default:
		return 0, 0, false, fmt.Errorf("%s is not a supported alloc_lb_strategy", s.cfg.AllocLBStrategy)
	}

	return s.allocClient.Free(ctx, addrs, namespace, resource, tokens, version)
}

func (s *Service) roundRobinLocked() []string {
	addrs := make([]string, 0, len(s.allocMembers))
	for _, instance := range s.allocMembers {
		addrs = append(addrs, net.JoinHostPort(instance.Host, strconv.Itoa(instance.HTTPPort)))
	}

	return addrs
}

func (s *Service) hashRingLocked(namespace, resource string) ([]string, error) {
	id := strings.Join([]string{namespace, resource}, "_")
	vNode, ok := s.rateHashRing.GetNode(id)
	if !ok {
		return nil, fmt.Errorf("failed to get address from ring: id=%s", id)
	}

	addr := s.trimVNodePrefix(vNode)

	return []string{addr}, nil
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
		s.updateRateMembersAndHashRing(members, 10)
	}

	members, err = s.fetchMembers(s.cfg.AllocAddresses)
	if err != nil {
		s.logger.Warn("failed to get alloc members", "err", err)
	} else {
		s.updateAllocMembersAndHashRing(members, 10)
	}
}

func (s *Service) fetchMembers(discoverAddrs []string) ([]domain.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	addresses, err := s.discoverer.Discover(ctx, discoverAddrs)
	if err != nil {
		return nil, fmt.Errorf("failed to discover: %w", err)
	}

	members, err := s.memberlistClient.Members(ctx, addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return members, nil
}

func (s *Service) updateRateMembersAndHashRing(newMembers []domain.Instance, numVNodes int) {
	s.logger.Info("updating rate hash ring", "new_members", fmt.Sprintf("%+v", newMembers))

	nodes := make([]string, 0, len(newMembers))
	for _, member := range newMembers {
		for i := 0; i < numVNodes; i++ {
			addr := fmt.Sprintf("%s:%d", member.Host, member.HTTPPort)
			vNode := s.prependVNodePrefix(addr, i)
			nodes = append(nodes, vNode)
		}
	}

	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	s.rateMembers = newMembers
	s.rateHashRing = hashring.New(nodes)
}

func (s *Service) updateAllocMembersAndHashRing(newMembers []domain.Instance, numVNodes int) {
	s.logger.Info("updating alloc hash ring", "new_members", fmt.Sprintf("%+v", newMembers))

	nodes := make([]string, 0, len(newMembers))
	for _, member := range newMembers {
		for i := 0; i < numVNodes; i++ {
			addr := fmt.Sprintf("%s:%d", member.Host, member.HTTPPort)
			vNode := s.prependVNodePrefix(addr, i)
			nodes = append(nodes, vNode)
		}
	}

	s.allocMu.Lock()
	defer s.allocMu.Unlock()

	s.allocMembers = newMembers
	s.allocHashRing = hashring.New(nodes)
}

func (s *Service) prependVNodePrefix(addr string, vNodeID int) string {
	return fmt.Sprintf("vnode%d_%s", vNodeID, addr)
}

func (s *Service) trimVNodePrefix(vNode string) string {
	before, addr, ok := strings.Cut(vNode, "_")

	s.logger.Info("trimVNodePrefix", "before", before, "addr", addr, "ok", ok)

	return addr
}
