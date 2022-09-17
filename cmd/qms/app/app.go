package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/middleware/gorillamux"
)

type App struct {
	cfg    Config
	logger log.Logger
	server *http.Server
}

func New(cfg Config, clock clock.Clock, logger log.Logger, reg prometheus.Registerer, tp trace.TracerProvider) (*App, error) {
	pingService := services.NewPingService()
	pingHTTPHandler := handlers.NewPingHTTPHandler(pingService)

	rateQuotaService, err := services.NewRateQuotaService(logger, clock, cfg.RateQuotaServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate quota service: %w", err)
	}

	rateQuotaHTTPHandler := handlers.NewRateQuotaHTTPHandler(rateQuotaService)

	allocationQuotaService, err := services.NewAllocationQuotaService(logger, cfg.AllocationQuotaServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocation quota service: %w", err)
	}

	allocationQuotaHTTPHandler := handlers.NewAllocationQuotaHTTPHandler(allocationQuotaService)

	router := mux.NewRouter()
	router.Use(
		gorillamux.TraceMiddleware(tp, "gorillamux"),
		gorillamux.MetricsMiddleware(clock, reg, "default", "qms", "gorillamux"),
		gorillamux.LogMiddleware(logger, "gorillamux"),
	)

	router.Handle("/metrics", promhttp.Handler())
	v1ApiRouter := router.PathPrefix("/api/v1").Subrouter()
	v1ApiRouter.HandleFunc("/ping", pingHTTPHandler.Ping()).Methods(http.MethodGet)
	v1ApiRouter.HandleFunc("/allow", rateQuotaHTTPHandler.Allow()).Methods(http.MethodPost)
	v1ApiRouter.HandleFunc("/alloc", allocationQuotaHTTPHandler.Alloc()).Methods(http.MethodPost)
	v1ApiRouter.HandleFunc("/free", allocationQuotaHTTPHandler.Free()).Methods(http.MethodPost)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      router,
		},
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting http server", zap.Int("port", a.cfg.HTTPPort))

	err := a.server.ListenAndServe()
	switch {
	case errors.Is(err, http.ErrServerClosed):
		return nil
	default:
	}

	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
