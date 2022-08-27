package app

import "github.com/Blinkuu/qms/internal/core/services"

type Config struct {
	HTTPPort           int                         `mapstructure:"http_port"`
	QuotaServiceConfig services.QuotaServiceConfig `mapstructure:"quota_service"`
}
