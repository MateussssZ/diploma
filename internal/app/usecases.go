package app

import (
	"apigateway/internal/clients"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/usecases"
)

type UsecasesDep struct {
	Repo          *Registries                   `validate:"required"`
	AuctionClient *clients.AuctionServiceClient `validate:"required"`
	JWTSecret     string                        `validate:"required"`
}

type Usecases struct {
	User    usecases.IUserUsecase
	Auction usecases.IAuctionUsecase
}

func NewUsecases(dep UsecasesDep) (*Usecases, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewUsecases", err)
	}

	user, err := usecases.NewUserUsecase(usecases.UserUsecaseDep{
		UserRepo:  dep.Repo.Postgres.User,
		JWTSecret: dep.JWTSecret,
	})
	if err != nil {
		return nil, err
	}

	auction, err := usecases.NewAuctionUsecase(usecases.AuctionUsecaseDep{
		AuctionClient: dep.AuctionClient,
	})
	if err != nil {
		return nil, err
	}

	return &Usecases{
		User:    user,
		Auction: auction,
	}, nil
}
