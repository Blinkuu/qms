package services

type PingService struct{}

func NewPingService() *PingService {
	return &PingService{}
}

func (s *PingService) Ping() string {
	return "pong"
}
