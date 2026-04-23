package config

import (
	"log/slog"
	"time"

	"apigateway/internal/pkg/validate"

	"github.com/spf13/viper"
)

const _defaultConfigPath = "config.yaml"

type (
	Logger struct {
		Level slog.Level `mapstructure:"Level" validate:"min=-4,max=8"`
	}

	GRPCServer struct {
		Address        string `mapstructure:"Address" validate:"required,min=1"`
		MaxRecvMsgSize int    `mapstructure:"MaxRecvMsgSize" validate:"gt=0"`
		MaxSendMsgSize int    `mapstructure:"MaxSendMsgSize" validate:"gt=0"`
	}

	RESTServer struct {
		ListenPort        string        `mapstructure:"ListenPort" validate:"required,min=1"`
		ReadTimeout       time.Duration `mapstructure:"ReadTimeout" validate:"gt=0"`
		WriteTimeout      time.Duration `mapstructure:"WriteTimeout" validate:"gt=0"`
		ShutdownTimeout   time.Duration `mapstructure:"ShutdownTimeout" validate:"gt=0"`
		ReadHeaderTimeout time.Duration `mapstructure:"ReadHeaderTimeout" validate:"gt=0"`
		IdleTimeout       time.Duration `mapstructure:"IdleTimeout" validate:"gt=0"`
		MaxHeaderBytes    int           `mapstructure:"MaxHeaderBytes" validate:"gt=0"`
	}
)

type Config struct {
	Logger         Logger     `mapstructure:"Logger" validate:"required"`
	GRPCServer     GRPCServer `mapstructure:"GRPCServer" validate:"required"`
	RESTServer     RESTServer `mapstructure:"RESTServer" validate:"required"`
	AuctionService GRPCClient `mapstructure:"AuctionService" validate:"required"`
	Kafka          Kafka      `mapstructure:"Kafka" validate:"required"`
}

// GRPCClient is the address of an external gRPC service.
type GRPCClient struct {
	Address string `mapstructure:"Address" validate:"required,min=1"`
}

// Kafka holds Kafka consumer configuration.
type Kafka struct {
	Brokers []string `mapstructure:"Brokers" validate:"required"`
	Topic   string   `mapstructure:"Topic" validate:"required"`
	GroupID string   `mapstructure:"GroupID" validate:"required"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigFile(_defaultConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}
