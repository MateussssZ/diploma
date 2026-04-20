package repo

import (
	"context"
	"time"

	repomodels "apigateway/internal/repo/models"
)

type IUserRepo interface {
	Register(ctx context.Context, login, email, passwordHash string) error
	FindByLogin(ctx context.Context, login string) (*repomodels.User, error)
	SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, token string) (int64, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
