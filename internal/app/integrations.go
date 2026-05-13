package app

import (
	"apigateway/config"
	"apigateway/internal/integrations/cache"
	"apigateway/internal/integrations/kafka"
	"apigateway/internal/integrations/wsmanager"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type IntegrationsDep struct {
	AuctionActions wsmanager.IAuctionActions `validate:"required"`
	Metrics        metrics.IMetrics          `validate:"required"`
	Logger         applogger.IAppLogger      `validate:"required"`
	KafkaCfg       config.Kafka
	CacheManager   *cache.CacheManager `validate:"required"`
}

type Integrations struct {
	WSManager     *wsmanager.WSManager
	KafkaConsumer *kafka.Consumer
	CacheManager  *cache.CacheManager
}

func NewIntegrations(dep IntegrationsDep) (*Integrations, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewIntegrations", err)
	}

	wsMgr := wsmanager.NewWSManager(dep.AuctionActions, dep.Metrics, dep.Logger)
	consumer := kafka.NewConsumer(dep.KafkaCfg, wsMgr, dep.CacheManager, dep.Metrics, dep.Logger)

	return &Integrations{
		WSManager:     wsMgr,
		KafkaConsumer: consumer,
		CacheManager:  dep.CacheManager,
	}, nil
}
