package native

import (
	"context"
	"strings"

	"github.com/thanos-io/thanos/pkg/discovery/dns"

	"github.com/Blinkuu/qms/pkg/log"
)

type Discoverer struct {
	logger      log.Logger
	dnsProvider *dns.Provider
}

func NewDiscoverer(logger log.Logger, dnsProvider *dns.Provider) *Discoverer {
	return &Discoverer{
		logger:      logger,
		dnsProvider: dnsProvider,
	}
}

func (d *Discoverer) Discover(ctx context.Context, serviceNames []string) ([]string, error) {
	if len(serviceNames) == 0 {
		return nil, nil
	}

	var ms, resolve []string

	for _, member := range serviceNames {
		if strings.Contains(member, "+") {
			resolve = append(resolve, member)
		} else {
			// No DNS SRV record to lookup, just append member
			ms = append(ms, member)
		}
	}

	err := d.dnsProvider.Resolve(ctx, resolve)
	if err != nil {
		d.logger.Error("failed to resolve members", "addrs", strings.Join(resolve, ","), "err", err)
	}

	ms = append(ms, d.dnsProvider.Addresses()...)

	return ms, nil
}
