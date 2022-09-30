package raft

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Join         bool   `yaml:"join"`
	DeploymentID uint64 `yaml:"deployment_id"`
	ReplicaID    uint64 `yaml:"replica_id"`
	ShardID      uint64 `yaml:"shard_id"`
	Dir          string `yaml:"dir"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.BoolVar(&c.Join, strutil.WithPrefixOrDefault(prefix, "join"), false, "")
	f.Uint64Var(&c.DeploymentID, strutil.WithPrefixOrDefault(prefix, "deployment_id"), 0, "")
	f.Uint64Var(&c.ReplicaID, strutil.WithPrefixOrDefault(prefix, "replica_id"), 0, "")
	f.Uint64Var(&c.ShardID, strutil.WithPrefixOrDefault(prefix, "shard_id"), 0, "")
	f.StringVar(&c.Dir, strutil.WithPrefixOrDefault(prefix, "dir"), "/tmp/qms/data/raft", "")
}
