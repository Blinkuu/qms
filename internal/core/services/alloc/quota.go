package alloc

import (
	"github.com/Blinkuu/qms/pkg/alloclimit"
)

type quota struct {
	Namespace string                    `mapstructure:"namespace"`
	Resource  string                    `mapstructure:"resource"`
	Strategy  alloclimit.StrategyConfig `mapstructure:"strategy"`
}
