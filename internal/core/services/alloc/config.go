package alloc

import (
	"flag"

	"github.com/Blinkuu/qms/internal/core/storage/alloc"
	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Quotas  quotaList    `yaml:"quotas"`
	Storage alloc.Config `yaml:"storage"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.Quotas, strutil.WithPrefixOrDefault(prefix, "quotas"), "")

	c.Storage.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "storage"))
}
