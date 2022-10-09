package domain

type Instance struct {
	Service    string `json:"service"`
	Hostname   string `json:"hostname"`
	Host       string `json:"host"`
	HTTPPort   int    `json:"http_port"`
	GossipPort int    `json:"gossip_port"`
}

func NewInstance(service, hostname, host string, httpPort, gossipPort int) Instance {
	return Instance{
		Service:    service,
		Hostname:   hostname,
		Host:       host,
		HTTPPort:   httpPort,
		GossipPort: gossipPort,
	}
}
