package proxy

import (
	"flag"

	"github.com/grafana/dskit/flagext"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	RateAddresses  flagext.StringSlice `yaml:"rate_addresses"`
	AllocAddresses flagext.StringSlice `yaml:"alloc_addresses"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.RateAddresses, strutil.WithPrefixOrDefault(prefix, "rate_addresses"), "")
	f.Var(&c.AllocAddresses, strutil.WithPrefixOrDefault(prefix, "alloc_addresses"), "")
}
