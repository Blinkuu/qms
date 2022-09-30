package domain

type Instance struct {
	Service    string `json:"service"`
	Host       string `json:"host"`
	HTTPPort   int    `json:"http_port"`
	GossipPort int    `json:"gossip_port"`
}

func NewInstance(service, host string, httpPort, gossipPort int) Instance {
	return Instance{
		Service:    service,
		Host:       host,
		HTTPPort:   httpPort,
		GossipPort: gossipPort,
	}
}
