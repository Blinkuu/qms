package alloc

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/storage/local"
	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Backend string       `yaml:"backend"`
	Local   local.Config `yaml:"local"`
	Quotas  quotaList    `yaml:"quotas"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Backend, strutil.WithPrefixOrDefault(prefix, "backend"), "local", "")
	f.Var(&c.Quotas, strutil.WithPrefixOrDefault(prefix, "quotas"), "")

	c.Local.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "local"))
}
