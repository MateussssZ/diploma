package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"authservice/internal/pkg/errorspkg"
	"authservice/internal/repo/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Register(ctx context.Context, login, email, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (login, email, password_hash) VALUES ($1, $2, $3)`,
		login, email, passwordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("UserRepo.Register: %w", errorspkg.NewUserAlreadyExistsError(login))
		}
		return fmt.Errorf("UserRepo.Register: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, login, email, password_hash FROM users WHERE login = $1`,
		login,
	)
	var u models.User
	if err := row.Scan(&u.ID, &u.Login, &u.Email, &u.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("UserRepo.FindByLogin: %w", errorspkg.NewUserNotFoundError(login))
		}
		return nil, fmt.Errorf("UserRepo.FindByLogin: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at`,
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("UserRepo.SaveRefreshToken: %w", err)
	}
	return nil
}

func (r *UserRepo) FindRefreshToken(ctx context.Context, token string) (int64, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens WHERE token = $1 AND expires_at > now()`,
		token,
	)
	var userID int64
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("UserRepo.FindRefreshToken: %w", errorspkg.NewTokenNotFoundError())
		}
		return 0, fmt.Errorf("UserRepo.FindRefreshToken: %w", err)
	}
	return userID, nil
}

func (r *UserRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("UserRepo.DeleteRefreshToken: %w", err)
	}
	return nil
}
