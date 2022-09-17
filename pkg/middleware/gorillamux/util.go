package gorillamux

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
)

type statusCodeRecordingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusCodeRecordingResponseWriter(w http.ResponseWriter) *statusCodeRecordingResponseWriter {
	return &statusCodeRecordingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // WriteHeader is not called if the response implicitly returns 200 OK, so we need to fill it here
	}
}

func (w *statusCodeRecordingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func httpFlavorFromRequest(r *http.Request) string {
	switch r.ProtoMajor {
	case 1:
		return fmt.Sprintf("1.%d", r.ProtoMinor)
	case 2:
		return "2"
	default:
		return "unknown"
	}
}

func httpHostFromRequest(r *http.Request) string {
	if h := r.Host; h != "" {
		return h
	}

	return "unknown"
}

func httpMethodFromRequest(r *http.Request) string {
	if m := r.Method; m != "" {
		return m
	}

	return "unknown"
}

func httpContentLengthFromRequest(r *http.Request) string {
	return strconv.FormatInt(r.ContentLength, 10)
}

func httpTargetFromRequest(r *http.Request) string {
	if t := r.RequestURI; t != "" {
		return t
	}

	return "unknown"
}

func httpRouteFromRequest(r *http.Request) string {
	var routeStr string
	route := mux.CurrentRoute(r)
	if route != nil {
		var err error
		routeStr, err = route.GetPathTemplate()
		if err != nil {
			routeStr, err = route.GetPathRegexp()
			if err != nil {
				routeStr = ""
			}
		}
	}
	if routeStr == "" {
		routeStr = "unknown"
	}

	return routeStr
}

func httpSchemeFromRequest(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	return "http"
}

func httpUserAgentFromRequest(r *http.Request) string {
	if ua := r.UserAgent(); ua != "" {
		return ua
	}

	return "unknown"
}

func traceIDFromContext(ctx context.Context) string {
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID()
	if !traceID.IsValid() {
		return "unknown"
	}

	traceIDStr := traceID.String()
	if traceIDStr == "" {
		return "unknown"
	}

	return traceIDStr
}
