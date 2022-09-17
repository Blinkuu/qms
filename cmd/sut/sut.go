package main

import (
	"net/http"

	"github.com/benbjohnson/clock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Blinkuu/qms/internal/core/services"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/middleware/gorillamux"
)

func main() {
	clk := clock.New()

	router := mux.NewRouter()
	router.Use(
		gorillamux.MetricsMiddleware(clk, prometheus.DefaultRegisterer, "default", "sut", "gorillamux"),
	)
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/api/v1/ping", handlers.NewPingHTTPHandler(services.NewPingService()).Ping())

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
