package app

import (
	"github.com/Blinkuu/qms/internal/core/services/alloc"
	"github.com/Blinkuu/qms/internal/core/services/rate"
	"github.com/Blinkuu/qms/internal/core/services/server"
)

type Config struct {
	Target              string        `mapstructure:"target"`
	OTelCollectorTarget string        `mapstructure:"otel_collector_target"`
	ServerConfig        server.Config `mapstructure:"server"`
	AllocServiceConfig  alloc.Config  `mapstructure:"alloc"`
	RateServiceConfig   rate.Config   `mapstructure:"rate"`
}
