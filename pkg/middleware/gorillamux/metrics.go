package gorillamux

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/benbjohnson/clock"
)

func MetricsMiddleware(clock clock.Clock, reg prometheus.Registerer, namespace, subsystem, serverName string) mux.MiddlewareFunc {
	// TODO(lukasz): Add request_size_bytes and response_size_bytes metrics
	reqTotal := promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name:      "requests_total",
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "The total number of requests received",
	}, []string{labelHTTPFlavor, labelHTTPHost, labelHTTPMethod, labelHTTPRoute, labelHTTPScheme, labelHTTPServerName, labelHTTPStatusCode, labelResult})

	reqDurationSeconds := promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
		Name:      "request_duration_seconds",
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "Histogram of the request duration",
		Buckets:   prometheus.DefBuckets,
	}, []string{labelHTTPFlavor, labelHTTPHost, labelHTTPMethod, labelHTTPRoute, labelHTTPScheme, labelHTTPServerName, labelHTTPStatusCode, labelResult})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := clock.Now()

			// TODO(lukasz): Use sync.Pool for recycling these objects
			sw := newStatusCodeRecordingResponseWriter(w)
			next.ServeHTTP(sw, r)

			labelValues := []string{
				httpFlavorFromRequest(r),
				httpHostFromRequest(r),
				httpMethodFromRequest(r),
				httpRouteFromRequest(r),
				httpSchemeFromRequest(r),
				serverName,
				strconv.Itoa(sw.statusCode),
				resultFromHTTPStatusCode(sw.statusCode),
			}

			reqTotal.WithLabelValues(labelValues...).Inc()
			reqDurationSeconds.WithLabelValues(labelValues...).Observe(clock.Now().Sub(start).Seconds())
		})
	}
}
