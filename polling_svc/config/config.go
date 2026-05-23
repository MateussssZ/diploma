package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	RESTServer     RESTServer
	AuctionService AuctionService
	AuthService    AuthService
}

type RESTServer struct {
	ListenPort      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type AuctionService struct {
	Address string
}

type AuthService struct {
	SocketPath string
	JWTSecret  string
}

func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetDefault("RESTServer.ListenPort", "8082")
	v.SetDefault("RESTServer.ReadTimeout", "30s")
	v.SetDefault("RESTServer.WriteTimeout", "30s")
	v.SetDefault("RESTServer.ShutdownTimeout", "5s")
	v.SetDefault("AuctionService.Address", "localhost:9090")
	v.SetDefault("AuthService.SocketPath", "C:/go/src/templateservice/authservice.sock")

	_ = v.ReadInConfig()

	return &Config{
		RESTServer: RESTServer{
			ListenPort:      v.GetString("RESTServer.ListenPort"),
			ReadTimeout:     v.GetDuration("RESTServer.ReadTimeout"),
			WriteTimeout:    v.GetDuration("RESTServer.WriteTimeout"),
			ShutdownTimeout: v.GetDuration("RESTServer.ShutdownTimeout"),
		},
		AuctionService: AuctionService{
			Address: v.GetString("AuctionService.Address"),
		},
		AuthService: AuthService{
			SocketPath: v.GetString("AuthService.SocketPath"),
			JWTSecret:  v.GetString("AuthService.JWTSecret"),
		},
	}, nil
}
