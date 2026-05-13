package cache

import (
	"context"
	"encoding/json"
	"time"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/pkg/applogger"
)

type CacheManager struct {
	redis      IRedisClient
	logger     applogger.IAppLogger
	ttls       CacheTTLs
	isDisabled bool
}

type CacheTTLs struct {
	AuctionDetailTTL time.Duration
}

func NewCacheManager(
	redis IRedisClient,
	logger applogger.IAppLogger,
	ttls CacheTTLs,
) *CacheManager {
	// If redis is nil, disable cache
	isDisabled := redis == nil

	return &CacheManager{
		redis:      redis,
		logger:     logger,
		ttls:       ttls,
		isDisabled: isDisabled,
	}
}

// GetAuctionDetail retrieves auction detail with caching using cache-aside pattern
func (cm *CacheManager) GetAuctionDetail(
	ctx context.Context,
	auctionID string,
	fetchFunc func() (*models.AuctionDetail, error),
) (*models.AuctionDetail, error) {
	// If cache is disabled, fetch directly
	if cm.isDisabled {
		return fetchFunc()
	}

	key := AuctionDetailKey(auctionID)

	// Try cache first
	if cached, err := cm.redis.Get(ctx, key); err == nil && cached != "" {
		var result *models.AuctionDetail
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			cm.logger.Debug(ctx, "cache hit", "key", key)
			return result, nil
		}
	}

	// Cache miss: fetch from service
	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(result); err == nil {
		_ = cm.redis.Set(ctx, key, string(data), cm.ttls.AuctionDetailTTL)
		cm.logger.Debug(ctx, "cache set", "key", key, "ttl", cm.ttls.AuctionDetailTTL.String())
	}

	return result, nil
}

// InvalidateAuctionCache removes all cache keys related to an auction
// Called when auction is modified (bid placed, status changed, etc)
func (cm *CacheManager) InvalidateAuctionCache(ctx context.Context, auctionID string) {
	if cm.isDisabled {
		return
	}

	keys := InvalidateAuctionKeys(auctionID)
	if err := cm.redis.Del(ctx, keys...); err != nil {
		cm.logger.Error(ctx, err,
			"auction_id", auctionID,
			"message", "failed to invalidate auction cache")
	} else {
		cm.logger.Debug(ctx, "invalidated auction cache", "auction_id", auctionID, "keys_count", len(keys))
	}
}

// Close closes the redis connection
func (cm *CacheManager) Close() error {
	if cm.isDisabled {
		return nil
	}
	return cm.redis.Close()
}
