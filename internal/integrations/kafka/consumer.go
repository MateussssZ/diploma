package kafka

import (
	"context"
	"encoding/json"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"apigateway/config"
	"apigateway/internal/integrations/cache"
	"apigateway/internal/integrations/wsmanager"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
)

// Consumer reads auction events from a Kafka topic and broadcasts them via WSManager.
type Consumer struct {
	cfg          config.Kafka
	wsManager    *wsmanager.WSManager
	cacheManager *cache.CacheManager
	metrics      metrics.IMetrics
	logger       applogger.IAppLogger
}

func NewConsumer(cfg config.Kafka, wsManager *wsmanager.WSManager, cacheManager *cache.CacheManager, m metrics.IMetrics, logger applogger.IAppLogger) *Consumer {
	return &Consumer{
		cfg:          cfg,
		wsManager:    wsManager,
		cacheManager: cacheManager,
		metrics:      m,
		logger:       logger,
	}
}

func (c *Consumer) newReader() *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        c.cfg.Brokers,
		Topic:          c.cfg.Topic,
		GroupID:        c.cfg.GroupID,
		CommitInterval: 0, // manual commit — at-least-once
	})
}

// Start runs the consume loop until ctx is cancelled (SIGINT / SIGTERM).
// On any broker error the reader is closed, a new one is created after backoff,
// and the loop resumes — no data is lost because offsets are committed manually.
func (c *Consumer) Start(ctx context.Context) error {
	const (
		backoffMin = 2 * time.Second
		backoffMax = 30 * time.Second
	)

	backoff := backoffMin
	reader := c.newReader()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			reader.Close()

			if ctx.Err() != nil {
				return nil
			}

			c.metrics.KafkaConsumerErrorsInc(c.cfg.Topic)
			c.logger.Info(ctx, "kafka fetch error, reconnecting",
				"error", err.Error(), "backoff", backoff.String())

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(backoff):
			}

			if backoff < backoffMax {
				backoff *= 2
			}

			reader = c.newReader()
			continue
		}

		// Connected — reset backoff
		backoff = backoffMin
		start := time.Now()

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Info(ctx, "failed to unmarshal kafka event", "error", err.Error())
			c.metrics.KafkaConsumerErrorsInc(c.cfg.Topic)
			_ = reader.CommitMessages(ctx, msg) // skip poison pill
			continue
		}

		c.handle(ctx, event)

		if err := reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				reader.Close()
				return nil
			}
			c.logger.Info(ctx, "kafka commit failed", "error", err.Error())
			c.metrics.KafkaConsumerErrorsInc(c.cfg.Topic)
		}

		c.metrics.KafkaMessagesConsumedInc(c.cfg.Topic)
		c.metrics.KafkaMessageProcessingDurationInc(c.cfg.Topic, time.Since(start).Seconds())
	}
}

func (c *Consumer) handle(ctx context.Context, event Event) {
	switch event.Type {
	case EventBidPlaced, EventAuctionStatusChanged:
		// Broadcast to WebSocket subscribers
		c.wsManager.Broadcast(event.AuctionID, wsmanager.WSEvent{
			Event:     event.Type,
			AuctionID: event.AuctionID,
			Payload:   event.Payload,
		})
		// Invalidate cache for this auction
		c.cacheManager.InvalidateAuctionCache(ctx, event.AuctionID)
	default:
		c.logger.Info(ctx, "unknown kafka event type", "type", event.Type)
	}
}
