package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	config Config
	logger *zap.Logger
	server *http.Server
}

func New(logger *zap.Logger, config Config) *App {
	pingService := services.NewPingService()
	pingHandler := handlers.NewHTTPHandler(pingService)

	router := mux.NewRouter()
	v1ApiRouter := router.PathPrefix("/api/v1").Subrouter()
	v1ApiRouter.HandleFunc("/ping", pingHandler.Ping()).Methods(http.MethodGet)

	return &App{
		config: config,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.HTTPPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      router,
		},
	}
}

func (a *App) Run() error {
	a.logger.Info("starting http server", zap.Int("port", a.config.HTTPPort))

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
