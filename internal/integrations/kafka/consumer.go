package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"apigateway/config"
	"apigateway/internal/integrations/wsmanager"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
)

type ICacheInvalidator interface {
	InvalidateAuctionCache(ctx context.Context, auctionID string)
}

// Consumer reads bid events from a Kafka topic and fans them out via WSManager.
type Consumer struct {
	cfg       config.Kafka
	wsManager *wsmanager.WSManager
	cache     ICacheInvalidator
	metrics   metrics.IMetrics
	logger    applogger.IAppLogger
}

func NewConsumer(cfg config.Kafka, wsManager *wsmanager.WSManager, cache ICacheInvalidator, m metrics.IMetrics, logger applogger.IAppLogger) *Consumer {
	return &Consumer{
		cfg:       cfg,
		wsManager: wsManager,
		cache:     cache,
		metrics:   m,
		logger:    logger,
	}
}

func (c *Consumer) newReader() *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        c.cfg.Brokers,
		Topic:          c.cfg.Topic,
		GroupID:        c.cfg.GroupID,
		CommitInterval: 0,                      // manual commit — at-least-once
		MinBytes:       10e3,                   // 10 KB — wait to accumulate before returning
		MaxBytes:       10e6,                   // 10 MB — max fetch size per request
		MaxWait:        500 * time.Millisecond, // max time to wait for MinBytes
		QueueCapacity:  100,                    // internal message buffer
	})
}

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

		var event BidEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Info(ctx, "failed to unmarshal kafka bid event", "error", err.Error())
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

func (c *Consumer) handle(ctx context.Context, event BidEvent) {
	auctionID := strconv.FormatInt(event.LotID, 10)

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, fmt.Errorf("kafka event handler panicked: %v", r),
				"auction_id", auctionID,
			)
		}
	}()

	c.logger.Info(ctx, "kafka bid event received",
		"auction_id", auctionID,
		"bid_id", event.BidID,
		"bidder_id", event.BidderID,
		"amount", event.Amount,
		"new_current_price", event.NewCurrentPrice,
	)

	// Invalidate the cache now that the bid is confirmed committed in PostgreSQL.
	if c.cache != nil {
		c.cache.InvalidateAuctionCache(ctx, auctionID)
	}

	// Broadcast bid_placed to all WebSocket subscribers.
	// This is the only source of WS fan-out — carries the full payload
	c.wsManager.Broadcast(auctionID, wsmanager.WSEvent{
		Event:     EventBidPlaced,
		AuctionID: auctionID,
		Payload: map[string]interface{}{
			"bid_id":            event.BidID,
			"bidder_id":         event.BidderID,
			"amount":            event.Amount,
			"new_current_price": event.NewCurrentPrice,
			"placed_at":         event.PlacedAt,
		},
	})
}
