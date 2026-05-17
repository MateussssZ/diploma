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
		EnablePprof       bool          `mapstructure:"EnablePprof"`
	}
)

type Config struct {
	Logger         Logger         `mapstructure:"Logger"`
	GRPCServer     GRPCServer     `mapstructure:"GRPCServer" validate:"required"`
	RESTServer     RESTServer     `mapstructure:"RESTServer" validate:"required"`
	AuctionService GRPCClient     `mapstructure:"AuctionService" validate:"required"`
	AuthService    AuthServiceCfg `mapstructure:"AuthService" validate:"required"`
	Redis          RedisConfig    `mapstructure:"Redis" validate:"required"`
	CacheConfig    CacheConfig    `mapstructure:"CacheConfig" validate:"required"`
	Kafka          Kafka          `mapstructure:"Kafka" validate:"required"`
	WS             WSConfig       `mapstructure:"WS" validate:"required"`
}

// GRPCClient is the address of an external gRPC service (TCP).
type GRPCClient struct {
	Address string `mapstructure:"Address" validate:"required,min=1"`
}

// AuthServiceCfg holds the unix-socket path for the AuthService.
type AuthServiceCfg struct {
	SocketPath string `mapstructure:"SocketPath" validate:"required,min=1"`
}

// Kafka holds Kafka consumer configuration.
type Kafka struct {
	Brokers []string `mapstructure:"Brokers" validate:"required"`
	Topic   string   `mapstructure:"Topic" validate:"required"`
	GroupID string   `mapstructure:"GroupID" validate:"required"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Address      string        `mapstructure:"Address" validate:"required,min=1"`
	DB           int           `mapstructure:"DB" validate:"gte=0,lte=15"`
	Password     string        `mapstructure:"Password"`
	MaxRetries   int           `mapstructure:"MaxRetries" validate:"gte=0"`
	PoolSize     int           `mapstructure:"PoolSize" validate:"gt=0"`
	DialTimeout  time.Duration `mapstructure:"DialTimeout" validate:"gt=0"`
	ReadTimeout  time.Duration `mapstructure:"ReadTimeout" validate:"gt=0"`
	WriteTimeout time.Duration `mapstructure:"WriteTimeout" validate:"gt=0"`
}

// CacheConfig holds cache TTL configuration.
type CacheConfig struct {
	AuctionDetailTTL time.Duration `mapstructure:"AuctionDetailTTL" validate:"gt=0"`
}

// WSConfig holds WebSocket manager tuning parameters.
// These directly control throughput of real-time bid/auction events.
//
//	NumActionWorkers: goroutines executing blocking gRPC calls (PlaceBid, CreateAuction).
//	  Formula: target_bid_rps / (1000ms / grpc_p99_latency_ms)
//	  Example: 500 bids/sec @ 20ms p99 → 500/(1000/20) = 10 workers
//	  Default: 4 (safe baseline; profile with pprof and adjust)
//
//	NumBroadcastWorkers: goroutines draining the broadcast channel.
//	  Each worker calls deliver() which fans out to subscribers.
//	  Scale if ws_broadcast_dropped_total rises under load.
//
//	BroadcastBufSize: shared channel capacity (events). If full, events are dropped
//	  and ws_broadcast_dropped_total increments.
//
//	ActionQueueSize: per-worker-pool job queue depth. If full, gRPC actions are
//	  dropped and clients receive an error WS event.
type WSConfig struct {
	NumActionWorkers    int `mapstructure:"NumActionWorkers" validate:"gt=0"`
	NumBroadcastWorkers int `mapstructure:"NumBroadcastWorkers" validate:"gt=0"`
	BroadcastBufSize    int `mapstructure:"BroadcastBufSize" validate:"gt=0"`
	ActionQueueSize     int `mapstructure:"ActionQueueSize" validate:"gt=0"`
	MaxConnections      int `mapstructure:"MaxConnections" validate:"gt=0"`
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
