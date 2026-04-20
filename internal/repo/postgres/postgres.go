package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // что-то там goose не нравится
	"github.com/pressly/goose/v3"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/utils"
)

const _defaultMigrationDir = "migrations"

func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errorspkg.NewInitError("postgres pool", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, errorspkg.NewInitError("postgres ping", err)
	}

	if err = utils.ApplyMigrations(ctx, goose.DialectPostgres, dsn, _defaultMigrationDir); err != nil {
		return nil, err
	}

	return pool, nil
}
