package ports

type PingService interface {
	Ping() string
}
