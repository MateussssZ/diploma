package app

import (
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
)

// Registries is kept as a placeholder in case direct DB access is needed in future.
type RepoDep struct{}

type Registries struct{}

func NewRepo(_ RepoDep) (*Registries, error) {
	if err := validate.Struct(struct{}{}); err != nil {
		return nil, errorspkg.NewValidationError("NewRepo", err)
	}
	return &Registries{}, nil
}
