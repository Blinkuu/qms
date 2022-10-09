package rate

import (
	"flag"

	"github.com/Blinkuu/qms/internal/core/storage/rate"
	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Quotas  quotaList   `yaml:"quotas"`
	Storage rate.Config `yaml:"storage"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.Quotas, strutil.WithPrefixOrDefault(prefix, "quotas"), "")

	c.Storage.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "storage"))
}
