package gorillamux

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Blinkuu/qms/pkg/log"
)

func LogMiddleware(logger log.Logger, serverName string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO(lukasz): Use sync.Pool for recycling these objects
			sw := newStatusCodeRecordingResponseWriter(w)
			next.ServeHTTP(sw, r)

			logger.Info(
				"incoming request",
				labelHTTPFlavor, httpFlavorFromRequest(r),
				labelHTTPHost, httpHostFromRequest(r),
				labelHTTPMethod, httpMethodFromRequest(r),
				labelHTTPRequestContentLength, httpContentLengthFromRequest(r),
				labelHTTPRoute, httpRouteFromRequest(r),
				labelHTTPScheme, httpSchemeFromRequest(r),
				labelHTTPServerName, serverName,
				labelHTTPStatusCode, strconv.Itoa(sw.statusCode),
				labelHTTPTarget, httpTargetFromRequest(r),
				labelHTTPUserAgent, httpUserAgentFromRequest(r),
				labelTraceID, traceIDFromContext(r.Context()),
			)
		})
	}
}
