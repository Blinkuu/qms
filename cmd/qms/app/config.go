package app

import (
	"flag"

	"github.com/Blinkuu/qms/internal/core/services/alloc"
	"github.com/Blinkuu/qms/internal/core/services/memberlist"
	"github.com/Blinkuu/qms/internal/core/services/rate"
	"github.com/Blinkuu/qms/internal/core/services/server"
	"github.com/Blinkuu/qms/pkg/strutil"
)

type Config struct {
	Target              string            `yaml:"target"`
	OTelCollectorTarget string            `yaml:"otel_collector_target"`
	ServerConfig        server.Config     `yaml:"server"`
	MemberlistConfig    memberlist.Config `yaml:"memberlist"`
	AllocConfig         alloc.Config      `yaml:"alloc"`
	RateConfig          rate.Config       `yaml:"rate"`
}

func (c *Config) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Target, strutil.WithPrefixOrDefault(prefix, "target"), SingleBinary, "")
	f.StringVar(&c.OTelCollectorTarget, strutil.WithPrefixOrDefault(prefix, "otel_collector_target"), "grafana-agent-traces:4317", "")

	c.ServerConfig.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "server"))
	c.MemberlistConfig.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "memberlist"))
	c.AllocConfig.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "alloc"))
	c.RateConfig.RegisterFlagsWithPrefix(f, strutil.WithPrefixOrDefault(prefix, "rate"))
}
