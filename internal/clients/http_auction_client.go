package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"apigateway/internal/api/rest/handlers/models"
)

// HTTPAuctionServiceClient is an HTTP-based client for the AuctionService.
type HTTPAuctionServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPAuctionServiceClient(baseURL string) *HTTPAuctionServiceClient {
	return &HTTPAuctionServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 0, // no timeout, let context handle it
		},
	}
}

// Lot represents an auction lot from the AuctionService
type Lot struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	StartingPrice float64 `json:"startingPrice"`
	CurrentPrice  float64 `json:"currentPrice"`
	Status        string  `json:"status"`
	ImageURL      string  `json:"imageUrl"`
	SellerID      int64   `json:"sellerId"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	EndsAt        string  `json:"endsAt"`
}

// PageResponse wraps paginated results
type PageResponse struct {
	Content       []Lot `json:"content"`
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	TotalElements int64 `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
	First         bool  `json:"first"`
	Last          bool  `json:"last"`
}

type CreateLotRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	StartingPrice float64 `json:"startingPrice"`
	Status        string  `json:"status"`
	ImageURL      string  `json:"imageUrl"`
	SellerID      int64   `json:"sellerId"`
	EndsAt        string  `json:"endsAt"`
}

type BidRequest struct {
	BidderID int64   `json:"bidderId"`
	Amount   float64 `json:"amount"`
}

type BidResponse struct {
	ID        int64   `json:"id"`
	LotID     int64   `json:"lotId"`
	BidderID  int64   `json:"bidderId"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"createdAt"`
}

type CurrentPriceResponse struct {
	LotID        int64   `json:"lotId"`
	CurrentPrice float64 `json:"currentPrice"`
}

// GetLots fetches all lots with pagination
func (c *HTTPAuctionServiceClient) GetLots(ctx context.Context, page int, size int) (*PageResponse, error) {
	url := fmt.Sprintf("%s/api/lots?page=%d&size=%d", c.baseURL, page, size)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result PageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetLot fetches a single lot by ID
func (c *HTTPAuctionServiceClient) GetLot(ctx context.Context, lotID int64) (*Lot, error) {
	url := fmt.Sprintf("%s/api/lots/%d", c.baseURL, lotID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var lot Lot
	if err := json.Unmarshal(body, &lot); err != nil {
		return nil, err
	}

	return &lot, nil
}

// CreateLot creates a new lot
func (c *HTTPAuctionServiceClient) CreateLot(ctx context.Context, req CreateLotRequest) (int64, error) {
	url := fmt.Sprintf("%s/api/lots", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var lot Lot
	if err := json.Unmarshal(body, &lot); err != nil {
		return 0, err
	}

	return lot.ID, nil
}

// PlaceBid places a bid on a lot
func (c *HTTPAuctionServiceClient) PlaceBid(ctx context.Context, lotID int64, amount float64, bidderId int64) (*models.PlaceBidResponse, error) {
	url := fmt.Sprintf("%s/api/bids/%d", c.baseURL, lotID)

	req := BidRequest{
		BidderID: bidderId,
		Amount:   amount,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var bidResp BidResponse
	if err := json.Unmarshal(body, &bidResp); err != nil {
		return nil, err
	}

	return &models.PlaceBidResponse{Success: true}, nil
}
