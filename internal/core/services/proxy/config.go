package proxy

import (
	"flag"

	"github.com/grafana/dskit/flagext"

	"github.com/Blinkuu/qms/pkg/strutil"
)

const (
	HashRingLBStrategy   = "hash-ring"
	RoundRobinLBStrategy = "round-robin"
)

type Config struct {
	RateAddresses   flagext.StringSlice `yaml:"rate_addresses"`
	AllocLBStrategy string              `yaml:"alloc_lb_strategy"`
	AllocAddresses  flagext.StringSlice `yaml:"alloc_addresses"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.RateAddresses, strutil.WithPrefixOrDefault(prefix, "rate_addresses"), "")
	f.StringVar(&c.AllocLBStrategy, strutil.WithPrefixOrDefault(prefix, "alloc_lb_strategy"), HashRingLBStrategy, "")
	f.Var(&c.AllocAddresses, strutil.WithPrefixOrDefault(prefix, "alloc_addresses"), "")
}
