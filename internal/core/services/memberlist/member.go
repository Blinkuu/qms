package memberlist

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/memberlist"

	"github.com/Blinkuu/qms/internal/core/domain"
)

type member struct {
	Service    string
	Hostname   string
	HTTPPort   int
	GossipPort int
}

func newMember(service, hostname string, gossipPort, httpPort int) *member {
	return &member{
		Service:    service,
		Hostname:   hostname,
		HTTPPort:   httpPort,
		GossipPort: gossipPort,
	}
}

func newMemberFromString(str string) (*member, error) {
	split := strings.Split(str, "/")
	if len(split) != 4 {
		return nil, fmt.Errorf("failed to split: str=%str", str)
	}

	service, hostname, httpPort, gossipPort := split[0], split[1], split[2], split[3]

	parsedGossipPort, err := strconv.Atoi(gossipPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gossip port string to int: %w", err)
	}

	parsedHTTPPort, err := strconv.Atoi(httpPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert http port string to int: %w", err)
	}

	return &member{
		Service:    service,
		Hostname:   hostname,
		GossipPort: parsedGossipPort,
		HTTPPort:   parsedHTTPPort,
	}, nil
}

func (m *member) String() string {
	return fmt.Sprintf("%s/%s/%d/%d", m.Service, m.Hostname, m.HTTPPort, m.GossipPort)
}

func nodeToInstance(node *memberlist.Node) (domain.Instance, error) {
	member, err := newMemberFromString(node.Name)
	if err != nil {
		return domain.Instance{}, fmt.Errorf("failed to split member name: memberName=%s", node.Name)
	}

	return domain.NewInstance(member.Service, member.Hostname, node.Addr.String(), member.HTTPPort, member.GossipPort), nil
}
