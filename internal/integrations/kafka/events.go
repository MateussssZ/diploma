package kafka

const (
	EventBidPlaced = "bid_placed"
)

//	{"bidId":1,"lotId":2,"bidderId":9,"amount":110000,"newCurrentPrice":110000,"placedAt":"2026-05-16T12:54:51"}
type BidEvent struct {
	BidID           int64  `json:"bidId"`
	LotID           int64  `json:"lotId"` // = auctionID in our domain
	BidderID        int64  `json:"bidderId"`
	Amount          int64  `json:"amount"`
	NewCurrentPrice int64  `json:"newCurrentPrice"`
	PlacedAt        string `json:"placedAt"`
}
