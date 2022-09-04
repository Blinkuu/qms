package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/log"
)

type App struct {
	cfg    Config
	clock  clock.Clock
	logger log.Logger
	server *http.Server
}

func New(clock clock.Clock, logger log.Logger, cfg Config) *App {
	pingService := services.NewPingService()
	pingHTTPHandler := handlers.NewPingHTTPHandler(pingService)

	quotaService := services.NewQuotaService(clock, logger, cfg.QuotaServiceConfig)
	quotaHTTPHandler := handlers.NewQuotaHTTPHandler(quotaService)

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	v1ApiRouter := router.PathPrefix("/api/v1").Subrouter()
	v1ApiRouter.HandleFunc("/ping", pingHTTPHandler.Ping()).Methods(http.MethodGet)
	v1ApiRouter.HandleFunc("/allow", quotaHTTPHandler.Allow()).Methods(http.MethodPost)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      router,
		},
	}
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
