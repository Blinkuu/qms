package rate

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Backend string    `yaml:"backend"`
	Quotas  quotaList `yaml:"quotas"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Backend, strutil.WithPrefixOrDefault(prefix, "backend"), "memory", "")
	f.Var(&c.Quotas, strutil.WithPrefixOrDefault(prefix, "quotas"), "")
}
