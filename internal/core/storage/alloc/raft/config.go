package raft

import (
	"flag"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	BindAddress             string `yaml:"bind_address"`
	BindPort                int    `yaml:"bind_port"`
	BindAddressFromHostname bool   `yaml:"bind_address_from_hostname"`
	DeploymentID            uint64 `yaml:"deployment_id"`
	ReplicaID               uint64 `yaml:"replica_id"`
	ReplicaIDOverride       string `yaml:"replica_id_override"`
	ShardID                 uint64 `yaml:"shard_id"`
	Shards                  uint64 `yaml:"shards"`
	Dir                     string `yaml:"dir"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.BindAddress, strutil.WithPrefixOrDefault(prefix, "bind_address"), "0.0.0.0", "")
	f.BoolVar(&c.BindAddressFromHostname, strutil.WithPrefixOrDefault(prefix, "bind_address_from_hostname"), false, "")
	f.IntVar(&c.BindPort, strutil.WithPrefixOrDefault(prefix, "bind_port"), 8832, "")
	f.Uint64Var(&c.DeploymentID, strutil.WithPrefixOrDefault(prefix, "deployment_id"), 1337, "")
	f.Uint64Var(&c.ReplicaID, strutil.WithPrefixOrDefault(prefix, "replica_id"), 1, "")
	f.StringVar(&c.ReplicaIDOverride, strutil.WithPrefixOrDefault(prefix, "replica_id_override"), "", "")
	f.Uint64Var(&c.ShardID, strutil.WithPrefixOrDefault(prefix, "shard_id"), 1, "")
	f.Uint64Var(&c.Shards, strutil.WithPrefixOrDefault(prefix, "shards"), 1, "")
	f.StringVar(&c.Dir, strutil.WithPrefixOrDefault(prefix, "dir"), "/tmp/qms/data/raft", "")
}
