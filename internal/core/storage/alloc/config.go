package alloc

import (
	"flag"

	"github.com/Blinkuu/qms/internal/core/storage/alloc/local"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/raft"
	"github.com/Blinkuu/qms/pkg/strutil"
)

const (
	Memory = "memory"
	Local  = "local"
	Raft   = "raft"
)

type Config struct {
	Backend string       `yaml:"backend"`
	Local   local.Config `yaml:"local"`
	Raft    raft.Config  `yaml:"raft"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Backend, strutil.WithPrefixOrDefault(prefix, "backend"), Memory, "")

	c.Local.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, Local))
	c.Raft.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, Raft))
}
