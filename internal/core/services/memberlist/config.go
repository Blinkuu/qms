package memberlist

import (
	"flag"
	"time"

	"github.com/grafana/dskit/flagext"

	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	JoinMembers    flagext.StringSlice `yaml:"join_members"`
	RejoinInterval time.Duration       `yaml:"rejoin_interval"`
	MinJoinBackoff time.Duration       `yaml:"min_join_backoff" `
	MaxJoinBackoff time.Duration       `yaml:"max_join_backoff"`
	MaxJoinRetries int                 `yaml:"max_join_retries"`
	LeaveTimeout   time.Duration       `yaml:"leave_timeout"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Var(&c.JoinMembers, strutil.WithPrefixOrDefault(prefix, "join_members"), "")
	f.DurationVar(&c.RejoinInterval, strutil.WithPrefixOrDefault(prefix, "rejoin_interval"), 0, "")
	f.DurationVar(&c.MinJoinBackoff, strutil.WithPrefixOrDefault(prefix, "min_join_backoff"), 1*time.Second, "")
	f.DurationVar(&c.MaxJoinBackoff, strutil.WithPrefixOrDefault(prefix, "max_join_backoff"), 30*time.Second, "")
	f.IntVar(&c.MaxJoinRetries, strutil.WithPrefixOrDefault(prefix, "max_join_retries"), 10, "")
	f.DurationVar(&c.LeaveTimeout, strutil.WithPrefixOrDefault(prefix, "leave_timeout"), 10*time.Second, "")
}
