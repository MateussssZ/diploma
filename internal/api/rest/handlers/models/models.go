package models

type RegisterRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessTokenExpiry int64  `json:"access_token_expiry"`
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessTokenExpiry int64  `json:"access_token_expiry"`
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
}

type AuctionBrief struct {
	AuctionID   string   `json:"auction_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PhotoURLs   []string `json:"photo_urls"`
	StartPrice  int64    `json:"start_price"`
	CurrentPrice int64   `json:"current_price"`
	EndTime     int64    `json:"end_time"`
	Status      string   `json:"status"`
	CreatorID   string   `json:"creator_id"`
}

type AuctionDetail struct {
	AuctionID    string   `json:"auction_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	PhotoURLs    []string `json:"photo_urls"`
	StartPrice   int64    `json:"start_price"`
	CurrentPrice int64    `json:"current_price"`
	EndTime      int64    `json:"end_time"`
	Status       string   `json:"status"`
	CreatorID    string   `json:"creator_id"`
	Bids         []Bid    `json:"bids"`
}

type Bid struct {
	BidderID  string `json:"bidder_id"`
	Amount    int64  `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}

type AuctionListResponse struct {
	Auctions   []AuctionBrief `json:"auctions"`
	Page       int32          `json:"page"`
	PageSize   int32          `json:"page_size"`
	TotalPages int32          `json:"total_pages"`
	TotalItems int32          `json:"total_items"`
}

type CreateAuctionRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"required,min=1,max=1000"`
	PhotoURLs   []string `json:"photo_urls" validate:"required,min=1"`
	StartPrice  int64    `json:"start_price" validate:"gt=0"`
	EndTime     int64    `json:"end_time" validate:"gt=0"`
}

type CreateAuctionResponse struct {
	AuctionID string `json:"auction_id"`
}

type SubscribeRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
}

type UnsubscribeRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
}

type PlaceBidRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
	Amount    int64  `json:"amount" validate:"gt=0"`
}

type PlaceBidResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AuctionUpdate struct {
	AuctionID    string   `json:"auction_id"`
	CurrentPrice int64    `json:"current_price"`
	LastBidTime  int64    `json:"last_bid_time"`
	Status       string   `json:"status"`
	Subscribers  []string `json:"subscribers"`
}
