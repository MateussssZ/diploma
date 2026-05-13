package errorspkg

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorHandler is the common interface for errors that carry HTTP and gRPC status codes.
// Both transport layers use errors.As(err, &h) to extract the right status.
type ErrorHandler interface {
	error
	HTTPStatus() int
	GRPCStatus() *status.Status
}

// UserNotFoundError is returned when a user lookup by login yields no result.
type UserNotFoundError struct {
	Login string
}

func NewUserNotFoundError(login string) *UserNotFoundError {
	return &UserNotFoundError{Login: login}
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user [%s] not found", e.Login)
}

func (e *UserNotFoundError) HTTPStatus() int {
	// Intentionally 401, not 404 — do not reveal whether the user exists.
	return http.StatusUnauthorized
}

func (e *UserNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.Unauthenticated, e.Error())
}

// UserAlreadyExistsError is returned on duplicate login/email during registration.
type UserAlreadyExistsError struct {
	Login string
}

func NewUserAlreadyExistsError(login string) *UserAlreadyExistsError {
	return &UserAlreadyExistsError{Login: login}
}

func (e *UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user [%s] already exists", e.Login)
}

func (e *UserAlreadyExistsError) HTTPStatus() int {
	return http.StatusConflict
}

func (e *UserAlreadyExistsError) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

// InvalidCredentialsError is returned when a password does not match the stored hash.
type InvalidCredentialsError struct{}

func NewInvalidCredentialsError() *InvalidCredentialsError {
	return &InvalidCredentialsError{}
}

func (e *InvalidCredentialsError) Error() string {
	return "invalid credentials"
}

func (e *InvalidCredentialsError) HTTPStatus() int {
	return http.StatusUnauthorized
}

func (e *InvalidCredentialsError) GRPCStatus() *status.Status {
	return status.New(codes.Unauthenticated, e.Error())
}

// TokenNotFoundError is returned when a refresh token is absent or has expired.
type TokenNotFoundError struct{}

func NewTokenNotFoundError() *TokenNotFoundError {
	return &TokenNotFoundError{}
}

func (e *TokenNotFoundError) Error() string {
	return "refresh token not found or expired"
}

func (e *TokenNotFoundError) HTTPStatus() int {
	return http.StatusUnauthorized
}

func (e *TokenNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.Unauthenticated, e.Error())
}
