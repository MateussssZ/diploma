package usecases

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"apigateway/internal/api/rest/handlers/models"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/repo"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type IUserUsecase interface {
	Register(ctx context.Context, req models.RegisterRequest) error
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error)
}

type UserUsecaseDep struct {
	UserRepo  repo.IUserRepo `validate:"required"`
	JWTSecret string         `validate:"required"`
}

type UserUsecase struct {
	userRepo  repo.IUserRepo
	jwtSecret []byte
}

func NewUserUsecase(dep UserUsecaseDep) (*UserUsecase, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewUserUsecase", err)
	}

	return &UserUsecase{
		userRepo:  dep.UserRepo,
		jwtSecret: []byte(dep.JWTSecret),
	}, nil
}

// Register hashes password and stores the new user.
func (u *UserUsecase) Register(ctx context.Context, req models.RegisterRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserUsecase.Register: bcrypt: %w", err)
	}

	if err := u.userRepo.Register(ctx, req.Login, req.Email, string(hash)); err != nil {
		return errorspkg.NewRepoError("UserUsecase", "Register", err)
	}

	return nil
}

// Login verifies credentials and returns access + refresh tokens.
func (u *UserUsecase) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := u.userRepo.FindByLogin(ctx, req.Login)
	if err != nil {
		return nil, errorspkg.NewRepoError("UserUsecase", "Login", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	accessExpiry := now.Add(accessTokenTTL)
	refreshExpiry := now.Add(refreshTokenTTL)

	accessToken, err := u.generateToken(user.ID, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Login: generate access token: %w", err)
	}

	refreshToken, err := u.generateToken(user.ID, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Login: generate refresh token: %w", err)
	}

	if err := u.userRepo.SaveRefreshToken(ctx, user.ID, refreshToken, refreshExpiry); err != nil {
		return nil, errorspkg.NewRepoError("UserUsecase", "Login.SaveRefreshToken", err)
	}

	return &models.LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpiry:  accessExpiry.Unix(),
		RefreshTokenExpiry: refreshExpiry.Unix(),
	}, nil
}

// Logout deletes the refresh token from the store.
func (u *UserUsecase) Logout(ctx context.Context, refreshToken string) error {
	if err := u.userRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return errorspkg.NewRepoError("UserUsecase", "Logout", err)
	}
	return nil
}

// Refresh validates the refresh token and issues new token pair.
func (u *UserUsecase) Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error) {
	userID, err := u.userRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errorspkg.NewRepoError("UserUsecase", "Refresh", err)
	}

	if err := u.userRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, errorspkg.NewRepoError("UserUsecase", "Refresh.DeleteOld", err)
	}

	now := time.Now()
	accessExpiry := now.Add(accessTokenTTL)
	refreshExpiry := now.Add(refreshTokenTTL)

	accessToken, err := u.generateToken(userID, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Refresh: generate access token: %w", err)
	}

	newRefreshToken, err := u.generateToken(userID, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Refresh: generate refresh token: %w", err)
	}

	if err := u.userRepo.SaveRefreshToken(ctx, userID, newRefreshToken, refreshExpiry); err != nil {
		return nil, errorspkg.NewRepoError("UserUsecase", "Refresh.SaveNew", err)
	}

	return &models.RefreshResponse{
		AccessToken:        accessToken,
		RefreshToken:       newRefreshToken,
		AccessTokenExpiry:  accessExpiry.Unix(),
		RefreshTokenExpiry: refreshExpiry.Unix(),
	}, nil
}

func (u *UserUsecase) generateToken(userID int64, expiry time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.jwtSecret)
}
