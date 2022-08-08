package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/env"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	pingService := services.NewPingService()
	pingHandler := handlers.NewHTTPHandler(pingService)

	router := mux.NewRouter()
	v1ApiRouter := router.PathPrefix("/api/v1").Subrouter()
	v1ApiRouter.HandleFunc("/ping", pingHandler.Ping()).Methods(http.MethodGet)

	s := &http.Server{
		Addr:         ":" + env.GetOrDefault("PRIMARY_PORT", "6789"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      router,
	}

	errChan := make(chan error)
	go func() {
		err := s.ListenAndServe()
		switch {
		case errors.Is(err, http.ErrServerClosed):
			errChan <- nil
		default:
		}

		errChan <- err
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	return <-errChan
}
