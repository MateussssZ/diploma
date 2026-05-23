package cache

import "fmt"

// Cache key patterns
const (
	AuctionDetailKeyPattern = "auction:detail:%s" // auctionID
)

// Key generator
func AuctionDetailKey(auctionID string) string {
	return fmt.Sprintf(AuctionDetailKeyPattern, auctionID)
}

// InvalidateAuctionKeys returns all cache keys related to an auction
func InvalidateAuctionKeys(auctionID string) []string {
	return []string{
		AuctionDetailKey(auctionID),
	}
}
