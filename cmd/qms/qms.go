package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Blinkuu/qms/cmd/qms/app"
	"github.com/Blinkuu/qms/pkg/log"
)

func main() {
	logger := log.Must(log.NewZapLogger("qms", "debug", []string{"stdout"}))
	defer func() {
		err := logger.Close()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			_, _ = fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
			os.Exit(1)
		}
	}()

	logger.Info("starting qms")

	cfg, err := loadConfig()
	if err != nil {
		logger.Panic("failed to load config", "err", err)
	}

	se, err := setupOpenTelemetryExporter(cfg.OTelCollectorTarget)
	if err != nil {
		logger.Panic("failed to setup open telemetry exporter", "err", err)
	}

	tp, err := setupTracerProvider(context.TODO(), se, "qms")
	if err != nil {
		logger.Panic("failed to setup tracer provider", "err", err)
	}

	a, err := app.New(cfg, clock.New(), logger, prometheus.DefaultRegisterer, tp)
	if err != nil {
		logger.Fatal("failed to create new app", "err", err)
	}

	if err := runAppAndWaitForSignal(a); err != nil {
		logger.Fatal("failed to shutdown qms server", zap.Error(err))
	}

	logger.Info("shutting down qms")
}

func runAppAndWaitForSignal(app *app.App) error {
	errChan := make(chan error)
	go func() {
		errChan <- app.Run()
		close(errChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown app: %w", err)
	}

	return <-errChan
}

func loadConfig() (app.Config, error) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return app.Config{}, fmt.Errorf("failed to read in config: %w", err)
	}

	var config app.Config
	if err := viper.Unmarshal(&config); err != nil {
		return app.Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func setupOpenTelemetryExporter(target string) (*otlptrace.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to opentelemetry collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	return traceExporter, nil
}

func setupTracerProvider(ctx context.Context, se tracesdk.SpanExporter, moduleName string) (trace.TracerProvider, error) {
	resources, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(moduleName),
		),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(se),
		tracesdk.WithResource(resources),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	return tp, nil
}
