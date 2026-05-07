package app

import (
	"apigateway/internal/clients"
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type ControllersDep struct {
	Usecases   *Usecases                  `validate:"required"`
	AuthClient *clients.AuthServiceClient `validate:"required"`
}

type Controllers struct {
	User    controllers.IUserCtrl
	Auction controllers.IAuctionCtrl
}

func NewControllers(dep ControllersDep) (*Controllers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewControllers", err)
	}

	auction, err := controllers.NewAuctionCtrl(controllers.AuctionCtrlDep{
		AuctionUsecase: dep.Usecases.Auction,
	})
	if err != nil {
		return nil, err
	}

	return &Controllers{
		User:    dep.AuthClient,
		Auction: auction,
	}, nil
}
