package rate

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

const (
	Memory = "memory"
)

type Config struct {
	Backend string `yaml:"backend"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Backend, strutil.WithPrefixOrDefault(prefix, "backend"), Memory, "")
}
