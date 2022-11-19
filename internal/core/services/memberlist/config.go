package memberlist

import (
	"flag"
	"time"

	"github.com/grafana/dskit/flagext"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	BindAddress      string              `yaml:"bind_address"`
	BindPort         int                 `yaml:"bind_port"`
	AdvertiseAddress string              `yaml:"advertise_address"`
	AdvertisePort    int                 `yaml:"advertise_port"`
	RejoinInterval   time.Duration       `yaml:"rejoin_interval"`
	MinJoinBackoff   time.Duration       `yaml:"min_join_backoff" `
	MaxJoinBackoff   time.Duration       `yaml:"max_join_backoff"`
	MaxJoinRetries   int                 `yaml:"max_join_retries"`
	LeaveTimeout     time.Duration       `yaml:"leave_timeout"`
	JoinAddresses    flagext.StringSlice `yaml:"join_addresses"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.BindAddress, strutil.WithPrefixOrDefault(prefix, "bind_address"), "0.0.0.0", "")
	f.IntVar(&c.BindPort, strutil.WithPrefixOrDefault(prefix, "bind_port"), 7946, "")
	f.StringVar(&c.AdvertiseAddress, strutil.WithPrefixOrDefault(prefix, "advertise_address"), "", "")
	f.IntVar(&c.AdvertisePort, strutil.WithPrefixOrDefault(prefix, "advertise_port"), 7946, "")
	f.DurationVar(&c.RejoinInterval, strutil.WithPrefixOrDefault(prefix, "rejoin_interval"), 0, "")
	f.DurationVar(&c.MinJoinBackoff, strutil.WithPrefixOrDefault(prefix, "min_join_backoff"), 1*time.Second, "")
	f.DurationVar(&c.MaxJoinBackoff, strutil.WithPrefixOrDefault(prefix, "max_join_backoff"), 30*time.Second, "")
	f.IntVar(&c.MaxJoinRetries, strutil.WithPrefixOrDefault(prefix, "max_join_retries"), 10, "")
	f.DurationVar(&c.LeaveTimeout, strutil.WithPrefixOrDefault(prefix, "leave_timeout"), 10*time.Second, "")
	f.Var(&c.JoinAddresses, strutil.WithPrefixOrDefault(prefix, "join_addresses"), "")
}
