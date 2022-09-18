package rate

import (
	"github.com/Blinkuu/qms/pkg/ratelimit"
)

type quota struct {
	Namespace string                   `mapstructure:"namespace"`
	Resource  string                   `mapstructure:"resource"`
	Strategy  ratelimit.StrategyConfig `mapstructure:"strategy"`
}
