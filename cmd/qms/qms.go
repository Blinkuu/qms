package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Blinkuu/qms/cmd/qms/app"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := zap.Must(zap.NewDevelopment())

	defer func() {
		if err := logger.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
			panic(err)
		}
	}()

	logger.Info("starting qms")

	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	a := app.New(logger, config)
	if err := runAppAndWaitForSignal(a); err != nil {
		logger.Error("failed to shutdown qms server", zap.Error(err))

		os.Exit(1)
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
