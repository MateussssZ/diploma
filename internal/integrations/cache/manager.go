package cache

import (
	"context"
	"encoding/json"
	"time"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/metrics"
	"apigateway/internal/pkg/applogger"
)

type CacheManager struct {
	redis      IRedisClient
	logger     applogger.IAppLogger
	metrics    metrics.IMetrics
	ttls       CacheTTLs
	isDisabled bool
}

type CacheTTLs struct {
	AuctionDetailTTL time.Duration
}

func NewCacheManager(
	redis IRedisClient,
	logger applogger.IAppLogger,
	m metrics.IMetrics,
	ttls CacheTTLs,
) *CacheManager {
	isDisabled := redis == nil
	return &CacheManager{
		redis:      redis,
		logger:     logger,
		metrics:    m,
		ttls:       ttls,
		isDisabled: isDisabled,
	}
}

// GetAuctionDetail retrieves auction detail with caching using cache-aside pattern.
func (cm *CacheManager) GetAuctionDetail(
	ctx context.Context,
	auctionID string,
	fetchFunc func() (*models.AuctionDetail, error),
) (*models.AuctionDetail, error) {
	if cm.isDisabled {
		return fetchFunc()
	}

	key := AuctionDetailKey(auctionID)

	if cached, err := cm.redis.Get(ctx, key); err == nil && cached != "" {
		var result *models.AuctionDetail
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			cm.logger.Debug(ctx, "cache hit", "key", key)
			cm.metrics.CacheHitInc("auction_detail")
			return result, nil
		}
		cm.metrics.CacheErrorInc("unmarshal")
	}

	cm.metrics.CacheMissInc("auction_detail")

	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		if setErr := cm.redis.Set(ctx, key, string(data), cm.ttls.AuctionDetailTTL); setErr != nil {
			cm.metrics.CacheErrorInc("set")
			cm.logger.Debug(ctx, "cache set failed", "key", key, "error", setErr.Error())
		} else {
			cm.logger.Debug(ctx, "cache set", "key", key, "ttl", cm.ttls.AuctionDetailTTL.String())
		}
	}

	return result, nil
}

// InvalidateAuctionCache removes all cache keys related to an auction.
func (cm *CacheManager) InvalidateAuctionCache(ctx context.Context, auctionID string) {
	if cm.isDisabled {
		return
	}

	keys := InvalidateAuctionKeys(auctionID)
	if err := cm.redis.Del(ctx, keys...); err != nil {
		cm.metrics.CacheErrorInc("del")
		cm.logger.Error(ctx, err, "auction_id", auctionID, "message", "failed to invalidate auction cache")
	} else {
		cm.metrics.CacheInvalidationInc("auction_detail")
		cm.logger.Debug(ctx, "invalidated auction cache", "auction_id", auctionID, "keys_count", len(keys))
	}
}

// Close closes the Redis connection pool.
func (cm *CacheManager) Close() error {
	if cm.isDisabled {
		return nil
	}
	return cm.redis.Close()
}
