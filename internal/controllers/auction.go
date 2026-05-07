package controllers

import (
	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/usecases"
	"context"
)

type IAuctionCtrl interface {
	GetAuctions(ctx context.Context, page, pageSize int) (*models.AuctionListResponse, error)
	GetAuctionByID(ctx context.Context, auctionID string) (*models.AuctionDetail, error)
	GetUserAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error)
	GetSubscribedAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error)
	CreateAuction(ctx context.Context, req models.CreateAuctionRequest, userID string) (string, error)
	PlaceBid(ctx context.Context, auctionID string, amount int64, userID string) (*models.PlaceBidResponse, error)
}

type AuctionCtrlDep struct {
	AuctionUsecase usecases.IAuctionUsecase `validate:"required"`
}

type AuctionCtrl struct {
	auctionUsecase usecases.IAuctionUsecase
}

func NewAuctionCtrl(dep AuctionCtrlDep) (*AuctionCtrl, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewAuctionCtrl", err)
	}

	return &AuctionCtrl{
		auctionUsecase: dep.AuctionUsecase,
	}, nil
}

func (c *AuctionCtrl) GetAuctions(ctx context.Context, page, pageSize int) (*models.AuctionListResponse, error) {
	return c.auctionUsecase.GetAuctions(ctx, page, pageSize)
}

func (c *AuctionCtrl) GetAuctionByID(ctx context.Context, auctionID string) (*models.AuctionDetail, error) {
	return c.auctionUsecase.GetAuctionByID(ctx, auctionID)
}

func (c *AuctionCtrl) GetUserAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error) {
	return c.auctionUsecase.GetUserAuctions(ctx, userID)
}

func (c *AuctionCtrl) GetSubscribedAuctions(ctx context.Context, userID string) (*models.AuctionListResponse, error) {
	return c.auctionUsecase.GetSubscribedAuctions(ctx, userID)
}

func (c *AuctionCtrl) CreateAuction(ctx context.Context, req models.CreateAuctionRequest, userID string) (string, error) {
	return c.auctionUsecase.CreateAuction(ctx, req, userID)
}

func (c *AuctionCtrl) PlaceBid(ctx context.Context, auctionID string, amount int64, userID string) (*models.PlaceBidResponse, error) {
	return c.auctionUsecase.PlaceBid(ctx, auctionID, amount, userID)
}
