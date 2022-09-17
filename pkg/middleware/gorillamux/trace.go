package gorillamux

import (
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
)

func TraceMiddleware(tp trace.TracerProvider, serverName string) mux.MiddlewareFunc {
	return otelmux.Middleware(serverName, otelmux.WithTracerProvider(tp))
}
