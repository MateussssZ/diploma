package usecases

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"authservice/internal/repo"
)

const (
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// TokenPair holds a freshly generated access + refresh token pair.
type TokenPair struct {
	AccessToken        string
	RefreshToken       string
	AccessTokenExpiry  int64
	RefreshTokenExpiry int64
}

type IUserUsecase interface {
	Register(ctx context.Context, login, password, email string) error
	Login(ctx context.Context, login, password string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type UserUsecaseDep struct {
	UserRepo   repo.IUserRepo
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type UserUsecase struct {
	userRepo   repo.IUserRepo
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewUserUsecase(dep UserUsecaseDep) (*UserUsecase, error) {
	if dep.UserRepo == nil {
		return nil, errors.New("UserUsecaseDep: UserRepo is required")
	}
	if dep.JWTSecret == "" {
		return nil, errors.New("UserUsecaseDep: JWTSecret is required")
	}

	accessTTL := dep.AccessTTL
	if accessTTL == 0 {
		accessTTL = defaultAccessTTL
	}
	refreshTTL := dep.RefreshTTL
	if refreshTTL == 0 {
		refreshTTL = defaultRefreshTTL
	}

	return &UserUsecase{
		userRepo:   dep.UserRepo,
		jwtSecret:  []byte(dep.JWTSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (u *UserUsecase) Register(ctx context.Context, login, password, email string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserUsecase.Register: bcrypt: %w", err)
	}
	return u.userRepo.Register(ctx, login, email, string(hash))
}

func (u *UserUsecase) Login(ctx context.Context, login, password string) (*TokenPair, error) {
	user, err := u.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return u.issueTokenPair(ctx, user.ID)
}

func (u *UserUsecase) Logout(ctx context.Context, refreshToken string) error {
	return u.userRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (u *UserUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := u.userRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("UserUsecase.Refresh: %w", err)
	}

	if err := u.userRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("UserUsecase.Refresh: delete old token: %w", err)
	}

	return u.issueTokenPair(ctx, userID)
}

// issueTokenPair generates a new access+refresh token pair and persists the refresh token.
func (u *UserUsecase) issueTokenPair(ctx context.Context, userID int64) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(u.accessTTL)
	refreshExpiry := now.Add(u.refreshTTL)

	accessToken, err := u.generateToken(userID, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair: access token: %w", err)
	}

	refreshToken, err := u.generateToken(userID, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("issueTokenPair: refresh token: %w", err)
	}

	if err := u.userRepo.SaveRefreshToken(ctx, userID, refreshToken, refreshExpiry); err != nil {
		return nil, fmt.Errorf("issueTokenPair: save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
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
