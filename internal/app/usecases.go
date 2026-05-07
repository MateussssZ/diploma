package app

import (
	"apigateway/internal/clients"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/usecases"
)

type UsecasesDep struct {
	AuctionClient *clients.AuctionServiceClient `validate:"required"`
}

type Usecases struct {
	Auction usecases.IAuctionUsecase
}

func NewUsecases(dep UsecasesDep) (*Usecases, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewUsecases", err)
	}

	auction, err := usecases.NewAuctionUsecase(usecases.AuctionUsecaseDep{
		AuctionClient: dep.AuctionClient,
	})
	if err != nil {
		return nil, err
	}

	return &Usecases{
		Auction: auction,
	}, nil
}
