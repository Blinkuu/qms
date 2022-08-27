package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	cfg    Config
	clock  clock.Clock
	logger *zap.Logger
	server *http.Server
}

func New(clock clock.Clock, logger *zap.Logger, cfg Config) *App {
	pingService := services.NewPingService()
	pingHTTPHandler := handlers.NewPingHTTPHandler(pingService)

	quotaService := services.NewQuotaService(clock, logger, cfg.QuotaServiceConfig)
	quotaHTTPHandler := handlers.NewQuotaHTTPHandler(quotaService)

	router := mux.NewRouter()
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
