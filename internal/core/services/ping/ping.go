package ping

import (
	"context"

	"github.com/grafana/dskit/services"

	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ServiceName = "ping"
)

type Service struct {
	services.NamedService
	logger log.Logger
}

func NewService(logger log.Logger) *Service {
	logger = logger.With("service", ServiceName)

	s := &Service{
		NamedService: nil,
		logger:       logger,
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s
}

func (s *Service) Ping(_ context.Context) string {
	return "pong"
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting ping service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running ping service")

	<-ctx.Done()

	return nil
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping ping service")

	if err != nil {
		s.logger.Error("ping service returned error from running state", "err", err)
	}

	return nil
}
