package models

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents user login response with tokens
type LoginResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessTokenExpiry int64  `json:"access_token_expiry"`
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
}

// LogoutRequest represents user logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents token refresh response
type RefreshResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessTokenExpiry int64  `json:"access_token_expiry"`
	RefreshTokenExpiry int64 `json:"refresh_token_expiry"`
}

// AuctionBrief represents brief auction information
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

// AuctionDetail represents detailed auction information
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

// Bid represents a bid in auction history
type Bid struct {
	BidderID  string `json:"bidder_id"`
	Amount    int64  `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}

// AuctionListResponse represents paged auction list response
type AuctionListResponse struct {
	Auctions   []AuctionBrief `json:"auctions"`
	Page       int32          `json:"page"`
	PageSize   int32          `json:"page_size"`
	TotalPages int32          `json:"total_pages"`
	TotalItems int32          `json:"total_items"`
}

// CreateAuctionRequest represents request to create a new auction
type CreateAuctionRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"required,min=1,max=1000"`
	PhotoURLs   []string `json:"photo_urls" validate:"required,min=1"`
	StartPrice  int64    `json:"start_price" validate:"gt=0"`
	EndTime     int64    `json:"end_time" validate:"gt=0"`
}

// CreateAuctionResponse represents response after creating auction
type CreateAuctionResponse struct {
	AuctionID string `json:"auction_id"`
}

// SubscribeRequest represents request to subscribe to auction
type SubscribeRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
}

// UnsubscribeRequest represents request to unsubscribe from auction
type UnsubscribeRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
}

// PlaceBidRequest represents request to place a bid
type PlaceBidRequest struct {
	AuctionID string `json:"auction_id" validate:"required"`
	Amount    int64  `json:"amount" validate:"gt=0"`
}

// PlaceBidResponse represents response after placing bid
type PlaceBidResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// AuctionUpdate represents auction update notification
type AuctionUpdate struct {
	AuctionID    string   `json:"auction_id"`
	CurrentPrice int64    `json:"current_price"`
	LastBidTime  int64    `json:"last_bid_time"`
	Status       string   `json:"status"`
	Subscribers  []string `json:"subscribers"`
}
