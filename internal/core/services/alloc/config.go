package alloc

import (
	"github.com/Blinkuu/qms/pkg/storage/local"
)

type Config struct {
	Backend string       `mapstructure:"backend"`
	Local   local.Config `mapstructure:"local"`
	Quotas  []quota      `mapstructure:"quotas"`
}
