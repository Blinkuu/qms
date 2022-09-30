package native

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/thanos-io/thanos/pkg/discovery/dns"

	"github.com/Blinkuu/qms/pkg/cloud"
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

func (d *Discoverer) Discover(ctx context.Context, serviceNames []string) ([]cloud.Instance, error) {
	if len(serviceNames) == 0 {
		return nil, nil
	}

	hostToInstance := make(map[string]cloud.Instance)
	var resolveHTTP []string
	var resolveGossip []string
	result := make([]cloud.Instance, 0)
	for _, serviceName := range serviceNames {
		if idx := strings.Index(serviceName, "+"); idx != -1 {
			resolveHTTP = append(resolveHTTP, fmt.Sprintf("%shttp.tcp.%s", serviceName[0:idx+1], serviceName[idx+1:]))
			resolveGossip = append(resolveGossip, fmt.Sprintf("%sgossip.tcp.%s", serviceName[0:idx+1], serviceName[idx+1:]))
		} else {
			host := serviceName
			port := 80
			splitHost, splitPort, err := net.SplitHostPort(serviceName)
			if err == nil {
				parsedPort, err := strconv.Atoi(splitPort)
				if err != nil {
					return nil, fmt.Errorf("failed to parse port string: %w", err)
				}

				host, port = splitHost, parsedPort
			}

			result = append(result, cloud.Instance{Host: host, HTTPPort: port})
		}
	}

	if err := d.dnsProvider.Resolve(ctx, resolveHTTP); err != nil {
		return nil, fmt.Errorf("failed to resolve http names: %w", err)
	}

	for _, addr := range d.dnsProvider.Addresses() {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to split host port: %w", err)
		}

		instance, found := hostToInstance[host]
		if !found {
			instance = cloud.Instance{Host: host, HTTPPort: 0, GossipPort: 0}
		}

		parsedPort, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse port string: %w", err)
		}

		instance.HTTPPort = parsedPort
		hostToInstance[host] = instance
	}

	if err := d.dnsProvider.Resolve(ctx, resolveGossip); err != nil {
		return nil, fmt.Errorf("failed to resolve gossip addresses: %w", err)
	}

	for _, addr := range d.dnsProvider.Addresses() {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to split host port: %w", err)
		}

		instance, found := hostToInstance[host]
		if !found {
			instance = cloud.Instance{Host: host, HTTPPort: 0, GossipPort: 0}
		}

		parsedPort, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse port string: %w", err)
		}

		instance.GossipPort = parsedPort
		hostToInstance[host] = instance
	}

	for _, instance := range hostToInstance {
		// If either port is zero, instance is invalid
		if instance.HTTPPort == 0 || instance.GossipPort == 0 {
			continue
		}

		result = append(result, instance)
	}

	return result, nil
}
