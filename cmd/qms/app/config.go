package app

import "github.com/Blinkuu/qms/internal/core/services"

type Config struct {
	HTTPPort                     int                                   `mapstructure:"http_port"`
	AllocationQuotaServiceConfig services.AllocationQuotaServiceConfig `mapstructure:"allocation_quota_service"`
	RateQuotaServiceConfig       services.RateQuotaServiceConfig       `mapstructure:"rate_quota_service"`
}
