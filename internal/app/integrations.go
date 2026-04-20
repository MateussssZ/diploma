package app

import (
	// "apigateway/internal/integrations"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

type IntegrationsDep struct {
	// креды, конфиги, прочее
}

type Integrations struct {
	// интерфейсы взаимодействия с интеграциями
	// S3 integrations.IS3
}

func NewIntegrations(dep IntegrationsDep) (*Integrations, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewIntegrations", err)
	}

	// инициализация интеграций
	// integrations.NewS3(integrations.S3Dep) -> (integrations.S3, err)

	return &Integrations{
		// S3: s3,
	}, nil
}
