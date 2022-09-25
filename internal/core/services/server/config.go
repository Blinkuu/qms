package server

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	HTTPPort int `yaml:"http_port"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.IntVar(&c.HTTPPort, strutil.WithPrefixOrDefault(prefix, "http_port"), 6789, "")
}
