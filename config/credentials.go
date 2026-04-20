package config

import (
	"apigateway/internal/pkg/validate"
	"fmt"

	"github.com/spf13/viper"
)

const _defaultCredentialsPath = "credentials.yaml"

type IDSNBuilder interface {
	ToDSN() string
}

type (
	Postgres struct {
		User     string `mapstructure:"User" validate:"required"`
		Password string `mapstructure:"Password" validate:"required"`
		Host     string `mapstructure:"Host" validate:"required"`
		Port     int    `mapstructure:"Port" validate:"gt=0"`
		Database string `mapstructure:"Database" validate:"required"`
	}

	Sentry struct {
		Dsn              string  `mapstructure:"Dsn" validate:"required"`
		TracesSampleRate float64 `mapstructure:"TracesSampleRate" validate:"gt=0"`
	}

	Auth struct {
		JWTSecret string `mapstructure:"JWTSecret" validate:"required"`
	}
)

type Credentials struct {
	Postgres Postgres `mapstructure:"Postgres"`
	Sentry   Sentry   `mapstructure:"Sentry"`
	Auth     Auth     `mapstructure:"Auth"`
}

func NewCredentials() (*Credentials, error) {
	viper.SetConfigFile(_defaultCredentialsPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var credentials Credentials
	if err := viper.Unmarshal(&credentials); err != nil {
		return nil, err
	}

	if err := validate.Struct(credentials); err != nil {
		return nil, err
	}

	return &credentials, nil
}

func (p Postgres) ToDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Database,
	)
}
