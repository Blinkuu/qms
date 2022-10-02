package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"github.com/grafana/dskit/services"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/middleware/gorillamux"
)

const (
	ServiceName = "server"
)

type Service struct {
	services.NamedService
	cfg     Config
	logger  log.Logger
	waitFor func() []services.Service
	HTTP    *mux.Router
	server  *http.Server
}

func NewService(cfg Config, clock clock.Clock, logger log.Logger, reg prometheus.Registerer, tp trace.TracerProvider, waitFor func() []services.Service) *Service {
	logger = logger.With("service", ServiceName)

	router := mux.NewRouter()
	router.Use(
		gorillamux.TimeoutMiddleware(10*time.Second),
		gorillamux.TraceMiddleware(tp, "gorillamux"),
		gorillamux.MetricsMiddleware(clock, reg, "default", "qms", "gorillamux"),
		gorillamux.LogMiddleware(logger, "gorillamux"),
	)

	s := &Service{
		NamedService: nil,
		cfg:          cfg,
		logger:       logger,
		waitFor:      waitFor,
		HTTP:         router,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      router,
		},
	}

	s.NamedService = services.NewBasicService(s.start, s.run, s.stop).WithName(ServiceName)

	return s
}

func (s *Service) start(_ context.Context) error {
	s.logger.Info("starting server service")

	return nil
}

func (s *Service) run(ctx context.Context) error {
	s.logger.Info("running server service", zap.Int("port", s.cfg.HTTPPort))

	go func() {
		err := s.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to listen and serve: %w", err)
		}
	}()

	<-ctx.Done()

	return nil
}

func (s *Service) stop(err error) error {
	s.logger.Info("stopping server service")

	if err != nil {
		s.logger.Error("server service returned error from running state", "err", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, service := range s.waitFor() {
		describedSvc := services.DescribeService(service)
		s.logger.Infof("waiting for %s service to be terminated", describedSvc)
		if err := service.AwaitTerminated(context.Background()); err != nil {
			s.logger.Errorf("failed to await terminated for %s service", describedSvc)
		}
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return err
}
