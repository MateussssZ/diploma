package kafka

import "encoding/json"

const (
	EventBidPlaced            = "bid_placed"
	EventAuctionStatusChanged = "auction_status_changed"
)

// Event is a domain event received from Kafka.
// Timestamp is Unix milliseconds; events within a partition are ordered by it.
type Event struct {
	Type      string          `json:"type"`
	AuctionID string          `json:"auction_id"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}
