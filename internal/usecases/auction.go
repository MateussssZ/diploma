package usecases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	auctionpb "apigateway/api/grpc/auctionservice"
	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/clients"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

// grpcCallTimeout caps individual gRPC calls to AuctionService.
// The gRPC service config also enforces a timeout, but this provides a
// defence-in-depth layer for calls originating from HTTP handlers.
const grpcCallTimeout = 4 * time.Second

type IAuctionUsecase interface {
	GetAuctions(ctx context.Context, page, pageSize int) (*models.AuctionListResponse, error)
	GetAuctionByID(ctx context.Context, auctionID string) (*models.AuctionDetail, error)
	GetUserAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error)
	GetSubscribedAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error)
	CreateAuction(ctx context.Context, req models.CreateAuctionRequest, userID string) (string, error)
	PlaceBid(ctx context.Context, auctionID string, amount int64, userID string) (*models.PlaceBidResponse, error)
}

type AuctionUsecaseDep struct {
	AuctionClient *clients.AuctionServiceClient `validate:"required"`
}

type AuctionUsecase struct {
	auctionClient *clients.AuctionServiceClient
}

func NewAuctionUsecase(dep AuctionUsecaseDep) (*AuctionUsecase, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewAuctionUsecase", err)
	}
	return &AuctionUsecase{
		auctionClient: dep.AuctionClient,
	}, nil
}

func (u *AuctionUsecase) GetAuctions(ctx context.Context, page, pageSize int) (*models.AuctionListResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	resp, err := u.auctionClient.LotClient.GetLots(callCtx, &auctionpb.GetLotsRequest{
		Page: int32(page),
		Size: int32(pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("AuctionUsecase.GetAuctions: %w", err)
	}

	out := &models.AuctionListResponse{
		Page:       resp.Page,
		PageSize:   resp.Size,
		TotalPages: resp.TotalPages,
		TotalItems: int32(resp.TotalElements),
	}
	for _, lot := range resp.Content {
		out.Auctions = append(out.Auctions, models.AuctionBrief{
			AuctionID:    strconv.FormatInt(lot.Id, 10),
			Title:        lot.Title,
			CurrentPrice: parseDecimalPrice(lot.CurrentPrice),
			Status:       lot.Status.String(),
		})
	}
	return out, nil
}

func (u *AuctionUsecase) GetAuctionByID(ctx context.Context, auctionID string) (*models.AuctionDetail, error) {
	id, err := strconv.ParseInt(auctionID, 10, 64)
	if err != nil {
		return nil, errorspkg.NewValidationError("AuctionUsecase.GetAuctionByID", fmt.Errorf("invalid auctionID: %w", err))
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	lot, err := u.auctionClient.LotClient.GetLot(callCtx, &auctionpb.GetLotRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("AuctionUsecase.GetAuctionByID: %w", err)
	}

	return &models.AuctionDetail{
		AuctionID:    strconv.FormatInt(lot.Id, 10),
		Title:        lot.Title,
		Description:  lot.Description,
		StartPrice:   parseDecimalPrice(lot.StartingPrice),
		CurrentPrice: parseDecimalPrice(lot.CurrentPrice),
		Status:       lot.Status.String(),
		CreatorID:    strconv.FormatInt(lot.SellerId, 10),
	}, nil
}

func (u *AuctionUsecase) GetUserAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error) {
	// TODO: AuctionService пока не поддерживает фильтрацию по продавцу —
	// запрашиваем активные лоты как временная заглушка
	callCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	resp, err := u.auctionClient.LotClient.GetLots(callCtx, &auctionpb.GetLotsRequest{
		Page:           0,
		Size:           100,
		FilterByStatus: true,
		Status:         auctionpb.LotStatus_ACTIVE,
	})
	if err != nil {
		return nil, fmt.Errorf("AuctionUsecase.GetUserAuctions: %w", err)
	}

	out := &models.AuctionListResponse{
		Page:       resp.Page,
		PageSize:   resp.Size,
		TotalPages: resp.TotalPages,
		TotalItems: int32(resp.TotalElements),
	}
	for _, lot := range resp.Content {
		out.Auctions = append(out.Auctions, models.AuctionBrief{
			AuctionID:    strconv.FormatInt(lot.Id, 10),
			Title:        lot.Title,
			CurrentPrice: parseDecimalPrice(lot.CurrentPrice),
			Status:       lot.Status.String(),
		})
	}
	return out, nil
}

func (u *AuctionUsecase) GetSubscribedAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error) {
	// TODO: AuctionService пока не поддерживает подписки —
	// временная заглушка, возвращает активные лоты
	return u.GetAuctions(ctx, 0, 50)
}

func (u *AuctionUsecase) CreateAuction(ctx context.Context, req models.CreateAuctionRequest, userID string) (string, error) {
	sellerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return "", errorspkg.NewValidationError("AuctionUsecase.CreateAuction", fmt.Errorf("invalid userID: %w", err))
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	lot, err := u.auctionClient.LotClient.CreateLot(callCtx, &auctionpb.CreateLotRequest{
		Title:         req.Title,
		Description:   req.Description,
		StartingPrice: strconv.FormatInt(req.StartPrice, 10),
		Status:        auctionpb.LotStatus_ACTIVE,
		SellerId:      sellerID,
	})
	if err != nil {
		return "", fmt.Errorf("AuctionUsecase.CreateAuction: %w", err)
	}

	return strconv.FormatInt(lot.Id, 10), nil
}

func (u *AuctionUsecase) PlaceBid(ctx context.Context, auctionID string, amount int64, userID string) (*models.PlaceBidResponse, error) {
	lotID, err := strconv.ParseInt(auctionID, 10, 64)
	if err != nil {
		return nil, errorspkg.NewValidationError("AuctionUsecase.PlaceBid", fmt.Errorf("invalid auctionID: %w", err))
	}
	bidderID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errorspkg.NewValidationError("AuctionUsecase.PlaceBid", fmt.Errorf("invalid userID: %w", err))
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	_, err = u.auctionClient.BidClient.PlaceBid(callCtx, &auctionpb.PlaceBidRequest{
		LotId:    lotID,
		BidderId: bidderID,
		Amount:   strconv.FormatInt(amount, 10),
	})
	if err != nil {
		return nil, fmt.Errorf("AuctionUsecase.PlaceBid: %w", err)
	}

	return &models.PlaceBidResponse{Success: true}, nil
}

// parseDecimalPrice конвертирует строковую цену "1500.00" в int64 (в копейках/центах)
func parseDecimalPrice(s string) int64 {
	f, _ := strconv.ParseFloat(s, 64)
	return int64(f * 100)
}
