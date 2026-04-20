package errorspkg

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type ErrorHandler interface {
	error
	HTTPStatus() int
}

// UnitIsNilError error unit is nil
type UnitIsNilError struct {
	unit string
}

func NewUnitIsNilError(unit string) *UnitIsNilError {
	return &UnitIsNilError{unit: unit}
}

func (err *UnitIsNilError) Error() string {
	return fmt.Sprintf("[%s] is nil", err.unit)
}

// UnitIsMissedError error unit is missed
type UnitIsMissedError struct {
	unit string
}

func NewUnitIsMissedError(unit string) *UnitIsMissedError {
	return &UnitIsMissedError{unit: unit}
}

func (err *UnitIsMissedError) Error() string {
	return fmt.Sprintf("[%s] is missed", err.unit)
}

// ValidationError validation error
type ValidationError struct {
	method string
	err    error
}

func NewValidationError(method string, err error) *ValidationError {
	return &ValidationError{
		method: method,
		err:    err,
	}
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("[%s] method got validation error: %v", err.method, err.err)
}

// InitError validation error
type InitError struct {
	object string
	err    error
}

func NewInitError(object string, err error) *InitError {
	return &InitError{
		object: object,
		err:    err,
	}
}

func (err *InitError) Error() string {
	return fmt.Sprintf("failed to init [%s]: %v", err.object, err.err)
}

// RepoError repository error
type RepoError struct {
	usecase string
	method  string
	err     error
}

func NewRepoError(usecase, repoMethod string, err error) *RepoError {
	return &RepoError{
		usecase: usecase,
		method:  repoMethod,
		err:     err,
	}
}

func (err *RepoError) Error() string {
	return fmt.Sprintf("repo method [%s] failed in usecase [%s] error: %v", err.method, err.usecase, err.err)
}

func (err *RepoError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, err.Error())
}

func (err *RepoError) HTTPStatus() int {
	return http.StatusInternalServerError
}
