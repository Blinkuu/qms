package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a8m/envsubst"
	"github.com/benbjohnson/clock"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v2"

	"github.com/Blinkuu/qms/cmd/qms/app"
	"github.com/Blinkuu/qms/pkg/log"
)

func main() {
	logger := log.Must(log.NewZapLogger("qms", "info", []string{"stderr"}))
	defer func() {
		err := logger.Close()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			_, _ = fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
			os.Exit(1)
		}
	}()

	logger.Info("starting qms")
	defer logger.Info("shutting down qms")

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

	hostname, err := os.Hostname()
	if err != nil {
		logger.Panic("failed to get hostname", "err", err)
	}

	a, err := app.New(
		cfg,
		clock.New(),
		logger.With("hostname", hostname),
		prometheus.DefaultRegisterer,
		tp,
	)
	if err != nil {
		logger.Fatal("failed to create new app", "err", err)
	}

	if err := runAppAndWaitForSignal(a); err != nil {
		logger.Fatal("failed to run app and wait for signal", "err", err)
	}
}

func runAppAndWaitForSignal(app *app.App) error {
	errChan := make(chan error)
	go func() {
		errChan <- app.Run(context.Background())
		close(errChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var err error
	select {
	case <-sigChan:
	case err = <-errChan:
		return fmt.Errorf("failed to run app: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown app: %w", err)
	}

	return err
}

func loadConfig() (app.Config, error) {
	const (
		configFileOption      = "config.file"
		configExpandEnvOption = "config.expand-env"
	)

	var (
		configFile      string
		configExpandEnv bool
	)

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&configFile, configFileOption, "./configs/config.yaml", "")
	fs.BoolVar(&configExpandEnv, configExpandEnvOption, false, "")

	args := os.Args[1:]
	for len(args) > 0 {
		err := fs.Parse(args)
		fmt.Println(err)
		args = args[1:]
	}

	cfg := app.Config{}
	cfg.RegisterFlagsWithPrefix(flag.CommandLine, "")
	if configFile != "" {
		buff, err := os.ReadFile(configFile)
		if err != nil {
			return app.Config{}, fmt.Errorf("failed to read configFile %s: %w", configFile, err)
		}

		if configExpandEnv {
			s, err := envsubst.Bytes(buff)
			if err != nil {
				return app.Config{}, fmt.Errorf("failed to expand env vars from configFile %s: %w", configFile, err)
			}
			buff = s
		}

		err = yaml.UnmarshalStrict(buff, &cfg)
		if err != nil {
			return app.Config{}, fmt.Errorf("failed to parse configFile %s: %w", configFile, err)
		}
	}

	flagext.IgnoredFlag(flag.CommandLine, configFileOption, "Configuration file to load")
	flagext.IgnoredFlag(flag.CommandLine, configExpandEnvOption, "Whether to expand environment variables in config file")
	flag.Parse()

	return cfg, nil
}

func setupOpenTelemetryExporter(target string) (*otlptrace.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithBlock(),
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
