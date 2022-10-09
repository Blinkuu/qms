package local

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Dir string `yaml:"dir"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Dir, strutil.WithPrefixOrDefault(prefix, "dir"), "/tmp/qms/data/local", "")
}
