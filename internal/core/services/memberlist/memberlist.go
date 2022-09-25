package memberlist

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/services"
	"github.com/hashicorp/memberlist"
	"github.com/thanos-io/thanos/pkg/discovery/dns"

	"github.com/Blinkuu/qms/internal/core/domain/cloud"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "memberlist"
)

type Service struct {
	services.NamedService
	cfg         Config
	logger      log.Logger
	dnsProvider *dns.Provider
	memberlist  *memberlist.Memberlist
}

func NewService(cfg Config, logger log.Logger, eventDelegate EventDelegate, service string, httpPort int) (*Service, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to read hostname: %w", err)
	}

	listCfg := memberlist.DefaultLANConfig()
	listCfg.Events = eventDelegateAdapter{EventDelegate: eventDelegate}
	listCfg.Name = newMember(service, hostname, listCfg.BindPort, httpPort).String()
	list, err := memberlist.Create(listCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create memberlist: %w", err)
	}

	s := &Service{
		NamedService: nil,
		cfg:          cfg,
		logger:       logger,
		dnsProvider:  dns.NewProvider(logger.Simple(), nil, dns.GolangResolverType),
		memberlist:   list,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Members() ([]*cloud.Instance, error) {
	var result []*cloud.Instance
	for _, memberNode := range s.memberlist.Members() {
		instance, err := nodeToInstance(memberNode)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to instance: %w", err)
		}

		result = append(result, instance)
	}

	return result, nil
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting memberlist service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running memberlist service")

	ok := s.joinMembersOnStartup(ctx)
	if !ok {
		s.logger.Warn("failed to join members on startup")
	}

	var tickerChan <-chan time.Time
	if s.cfg.RejoinInterval > 0 && len(s.cfg.JoinMembers) > 0 {
		t := time.NewTicker(s.cfg.RejoinInterval)
		defer t.Stop()

		tickerChan = t.C
	}

	for {
		select {
		case <-tickerChan:
			members := s.discoverMembers(ctx, s.cfg.JoinMembers)

			reached, err := s.memberlist.Join(members)
			if err == nil {
				s.logger.Info("re-joined memberlist cluster", "reached_nodes", reached)
			} else {
				s.logger.Warn("failed to re-join memberlist cluster", "err", err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping memberlist service")

	if err != nil {
		s.logger.Error("memberlist service returned error from running state", "err", err)
	}

	err = s.memberlist.Leave(s.cfg.LeaveTimeout)
	if err != nil {
		s.logger.Error("failed to leave memberlist", "err", err)
	}

	err = s.memberlist.Shutdown()
	if err != nil {
		s.logger.Error("failed to shutdown memberlist", "err", err)
	}

	return nil
}

func (s *Service) joinMembersOnStartup(ctx context.Context) bool {
	if len(s.cfg.JoinMembers) == 0 {
		return true
	}

	startTime := time.Now()

	s.logger.Info("joining memberlist cluster", "join_members", strings.Join(s.cfg.JoinMembers, ","))

	cfg := backoff.Config{
		MinBackoff: s.cfg.MinJoinBackoff,
		MaxBackoff: s.cfg.MaxJoinBackoff,
		MaxRetries: s.cfg.MaxJoinRetries,
	}

	boff := backoff.New(ctx, cfg)
	var lastErr error

	for boff.Ongoing() {
		// We rejoin all nodes, including those that were joined during "fast-join".
		// This is harmless and simpler.
		nodes := s.discoverMembers(ctx, s.cfg.JoinMembers)

		if len(nodes) > 0 {
			reached, err := s.memberlist.Join(nodes) // err is only returned if reached==0.
			if err == nil {
				s.logger.Info("joining memberlist cluster succeeded", "reached_nodes", reached, "elapsed_time", time.Since(startTime))

				return true
			}

			s.logger.Warn("joining memberlist cluster: failed to reach any nodes", "retries", boff.NumRetries(), "err", err)
			lastErr = err
		} else {
			s.logger.Warn("joining memberlist cluster: found no nodes to join", "retries", boff.NumRetries())
		}

		boff.Wait()
	}

	s.logger.Error("joining memberlist cluster failed", "last_error", lastErr, "elapsed_time", time.Since(startTime))

	return false
}

func (s *Service) discoverMembers(ctx context.Context, members []string) []string {
	if len(members) == 0 {
		return nil
	}

	var result, resolve []string
	for _, member := range members {
		if strings.Contains(member, "+") {
			resolve = append(resolve, member)
		} else {
			// No DNS SRV record to lookup, just append member
			result = append(result, member)
		}
	}

	err := s.dnsProvider.Resolve(ctx, resolve)
	if err != nil {
		s.logger.Warn("failed to resolve members", "addrs", strings.Join(resolve, ","), "err", err)
	}

	result = append(result, s.dnsProvider.Addresses()...)

	return result
}
