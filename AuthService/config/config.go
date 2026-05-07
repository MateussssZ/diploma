package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

const _defaultConfigPath = "config.yaml"

type (
	Logger struct {
		Level slog.Level `mapstructure:"Level"`
	}

	GRPCServer struct {
		SocketPath     string `mapstructure:"SocketPath"`
		MaxRecvMsgSize int    `mapstructure:"MaxRecvMsgSize"`
		MaxSendMsgSize int    `mapstructure:"MaxSendMsgSize"`
	}

	Postgres struct {
		User     string `mapstructure:"User"`
		Password string `mapstructure:"Password"`
		Host     string `mapstructure:"Host"`
		Port     int    `mapstructure:"Port"`
		Database string `mapstructure:"Database"`
	}

	Auth struct {
		JWTSecret  string        `mapstructure:"JWTSecret"`
		AccessTTL  time.Duration `mapstructure:"AccessTTL"`
		RefreshTTL time.Duration `mapstructure:"RefreshTTL"`
	}
)

func (p Postgres) ToDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

type Config struct {
	Logger     Logger     `mapstructure:"Logger"`
	GRPCServer GRPCServer `mapstructure:"GRPCServer"`
	Postgres   Postgres   `mapstructure:"Postgres"`
	Auth       Auth       `mapstructure:"Auth"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigFile(_defaultConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
