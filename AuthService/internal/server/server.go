package server

import (
	"context"
	"errors"

	authpb "authservice/api/auth"
	"authservice/internal/pkg/errorspkg"
	"authservice/internal/usecases"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the gRPC AuthServiceServer interface.
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	usecase usecases.IUserUsecase
}

func NewAuthServer(uc usecases.IUserUsecase) *AuthServer {
	return &AuthServer{usecase: uc}
}

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if err := s.usecase.Register(ctx, req.GetLogin(), req.GetPassword(), req.GetEmail()); err != nil {
		return nil, grpcErr(err)
	}
	return &authpb.RegisterResponse{}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	pair, err := s.usecase.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, grpcErr(err)
	}
	return &authpb.LoginResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessTokenExpiry:  pair.AccessTokenExpiry,
		RefreshTokenExpiry: pair.RefreshTokenExpiry,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if err := s.usecase.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, grpcErr(err)
	}
	return &authpb.LogoutResponse{}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.RefreshResponse, error) {
	pair, err := s.usecase.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, grpcErr(err)
	}
	return &authpb.RefreshResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessTokenExpiry:  pair.AccessTokenExpiry,
		RefreshTokenExpiry: pair.RefreshTokenExpiry,
	}, nil
}

// grpcErr maps domain errors to gRPC status errors via the ErrorHandler interface.
// Falls back to codes.Internal for unexpected errors.
func grpcErr(err error) error {
	var h errorspkg.ErrorHandler
	if errors.As(err, &h) {
		return h.GRPCStatus().Err()
	}
	return status.Errorf(codes.Internal, "%v", err)
}
