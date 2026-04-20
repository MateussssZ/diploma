package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/pressly/goose/v3"
	"apigateway/internal/pkg/errorspkg"
	"log/slog"
)

func ApplyMigrations(ctx context.Context, dialect goose.Dialect, dsn string, migrationDir string) error {
	db, err := goose.OpenDBWithDriver(string(dialect), dsn)
	if err != nil {
		return errorspkg.NewInitError(fmt.Sprintf("%s migration", string(dialect)), err)
	}
	defer db.Close()

	if err = goose.UpContext(ctx, db, migrationDir); err != nil {
		if errors.Is(err, goose.ErrNoMigrationFiles) {
			slog.Info(fmt.Sprintf("%v: %s", err, migrationDir))
			return nil
		}
		
		return errorspkg.NewInitError(fmt.Sprintf("%s up migration", string(dialect)), err)
	}

	return nil
}
