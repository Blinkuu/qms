package main

import (
	"net/http"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Blinkuu/qms/internal/core/services/ping"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/log"
	"github.com/Blinkuu/qms/pkg/middleware/gorillamux"
)

func main() {
	clk := clock.New()

	pingService := ping.NewService(log.NewNoopLogger())
	pingHandler := handlers.NewPingHTTPHandler(pingService)

	router := mux.NewRouter()
	router.Use(
		gorillamux.MetricsMiddleware(clk, prometheus.DefaultRegisterer, "default", "sut", "gorillamux"),
	)
	router.Handle("/metrics", promhttp.Handler())
	v1ApiRouter := router.PathPrefix("/api/v1").Subrouter()
	v1ApiRouter.Handle("/ping", pingHandler.Ping())

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
