package app

import (
	"apigateway/internal/controllers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type ControllersDep struct {
	Usecases *Usecases `validate:"required"`
}

type Controllers struct {
	User    controllers.IUserCtrl
	Auction controllers.IAuctionCtrl
}

func NewControllers(dep ControllersDep) (*Controllers, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewControllers", err)
	}

	user, err := controllers.NewUserCtrl(controllers.UserCtrlDep{
		UserUsecase: dep.Usecases.User,
	})
	if err != nil {
		return nil, err
	}

	auction, err := controllers.NewAuctionCtrl(controllers.AuctionCtrlDep{
		AuctionUsecase: dep.Usecases.Auction,
		UserUsecase:    dep.Usecases.User,
	})
	if err != nil {
		return nil, err
	}

	return &Controllers{
		User:    user,
		Auction: auction,
	}, nil
}
