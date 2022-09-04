package services

import "context"

type PingService struct{}

func NewPingService() *PingService {
	return &PingService{}
}

func (s *PingService) Ping(_ context.Context) string {
	return "pong"
}
