package app

import (
	"apigateway/config"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"apigateway/internal/repo"
	"apigateway/internal/repo/postgres"
	"context"
)

type RepoDep struct {
	PostgresCfg config.IDSNBuilder `validate:"required"`
}

type PostgresRegistry struct {
	User repo.IUserRepo
}

type Registries struct {
	Postgres PostgresRegistry
}

func NewRepo(ctx context.Context, dep RepoDep) (*Registries, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewRepo", err)
	}

	postgresConn, err := postgres.New(ctx, dep.PostgresCfg.ToDSN())
	if err != nil {
		return nil, err
	}

	return &Registries{
		Postgres: PostgresRegistry{
			User: postgres.NewUserRepo(postgresConn),
		},
	}, nil
}
