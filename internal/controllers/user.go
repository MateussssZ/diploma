package controllers

import (
	"context"
	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/usecases"
	"apigateway/internal/pkg/validate"
)

type IUserCtrl interface {
	Register(ctx context.Context, req models.RegisterRequest) error
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error)
}

type UserCtrlDep struct {
	UserUsecase usecases.IUserUsecase `validate:"required"`
}

type UserCtrl struct {
	userUsecase usecases.IUserUsecase
}

func NewUserCtrl(dep UserCtrlDep) (*UserCtrl, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewUserCtrl", err)
	}

	return &UserCtrl{
		userUsecase: dep.UserUsecase,
	}, nil
}

func (c *UserCtrl) Register(ctx context.Context, req models.RegisterRequest) error {
	return c.userUsecase.Register(ctx, req)
}

func (c *UserCtrl) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	return c.userUsecase.Login(ctx, req)
}

func (c *UserCtrl) Logout(ctx context.Context, refreshToken string) error {
	return c.userUsecase.Logout(ctx, refreshToken)
}

func (c *UserCtrl) Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error) {
	return c.userUsecase.Refresh(ctx, refreshToken)
}
