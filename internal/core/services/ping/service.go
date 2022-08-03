package ping

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Ping() string {
	return "Pong"
}
