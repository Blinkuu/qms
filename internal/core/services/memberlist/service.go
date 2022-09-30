package memberlist

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/services"
	"github.com/hashicorp/memberlist"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/pkg/cloud"
	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/strutil"
)

const (
	ServiceName = "memberlist"
)

type Config struct {
	JoinAddresses  flagext.StringSlice `yaml:"join_addresses"`
	RejoinInterval time.Duration       `yaml:"rejoin_interval"`
	MinJoinBackoff time.Duration       `yaml:"min_join_backoff" `
	MaxJoinBackoff time.Duration       `yaml:"max_join_backoff"`
	MaxJoinRetries int                 `yaml:"max_join_retries"`
	LeaveTimeout   time.Duration       `yaml:"leave_timeout"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.JoinAddresses, strutil.WithPrefixOrDefault(prefix, "join_addresses"), "")
	f.DurationVar(&c.RejoinInterval, strutil.WithPrefixOrDefault(prefix, "rejoin_interval"), 0, "")
	f.DurationVar(&c.MinJoinBackoff, strutil.WithPrefixOrDefault(prefix, "min_join_backoff"), 1*time.Second, "")
	f.DurationVar(&c.MaxJoinBackoff, strutil.WithPrefixOrDefault(prefix, "max_join_backoff"), 30*time.Second, "")
	f.IntVar(&c.MaxJoinRetries, strutil.WithPrefixOrDefault(prefix, "max_join_retries"), 10, "")
	f.DurationVar(&c.LeaveTimeout, strutil.WithPrefixOrDefault(prefix, "leave_timeout"), 10*time.Second, "")
}

type Service struct {
	services.NamedService
	cfg        Config
	logger     log.Logger
	discoverer cloud.Discoverer
	memberlist *memberlist.Memberlist
}

func NewService(cfg Config, logger log.Logger, discoverer cloud.Discoverer, eventDelegate EventDelegate, service string, httpPort int) (*Service, error) {
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
		discoverer:   discoverer,
		memberlist:   list,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s, nil
}

func (s *Service) Members(_ context.Context) ([]domain.Instance, error) {
	var result []domain.Instance
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
	if s.cfg.RejoinInterval > 0 && len(s.cfg.JoinAddresses) > 0 {
		t := time.NewTicker(s.cfg.RejoinInterval)
		defer t.Stop()

		tickerChan = t.C
	}

	for {
		select {
		case <-tickerChan:
			//addrs, err := s.discoverMembers(ctx, s.cfg.JoinAddresses)
			//if err != nil {
			//	s.logger.Warn("failed to discover members: %w", err)
			//	continue
			//}

			reached, err := s.memberlist.Join(s.cfg.JoinAddresses)
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
	if len(s.cfg.JoinAddresses) == 0 {
		return true
	}

	startTime := time.Now()

	s.logger.Info("joining memberlist cluster", "join_addresses", strings.Join(s.cfg.JoinAddresses, ","))

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
		//nodes, err := s.discoverMembers(ctx, s.cfg.JoinAddresses)
		//if err != nil {
		//	s.logger.Warn("failed to discover members", "err", err)
		//	continue
		//}

		if len(s.cfg.JoinAddresses) > 0 {
			reached, err := s.memberlist.Join(s.cfg.JoinAddresses) // err is only returned if reached==0.
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

//
//func (s *Service) discoverMembers(ctx context.Context, addrs []string) ([]string, error) {
//	if len(addrs) == 0 {
//		return nil, nil
//	}
//
//	instances, err := s.discoverer.Discover(ctx, addrs)
//	if err != nil {
//		return nil, fmt.Errorf("failed to discover: %w", err)
//	}
//
//	result := make([]string, 0, len(instances))
//	for _, instance := range instances {
//		addr := net.JoinHostPort(instance.Host, strconv.Itoa(instance.GossipPort))
//		result = append(result, addr)
//	}
//
//	s.logger.Info("discoverMembers", "result", result)
//
//	return result, nil
//}
