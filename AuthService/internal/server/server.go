package server

import (
	"context"

	authpb "authservice/api/auth"
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
		return nil, status.Errorf(codes.Internal, "register: %v", err)
	}
	return &authpb.RegisterResponse{}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	pair, err := s.usecase.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login: %v", err)
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
		return nil, status.Errorf(codes.Internal, "logout: %v", err)
	}
	return &authpb.LogoutResponse{}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.RefreshResponse, error) {
	pair, err := s.usecase.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh: %v", err)
	}
	return &authpb.RefreshResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessTokenExpiry:  pair.AccessTokenExpiry,
		RefreshTokenExpiry: pair.RefreshTokenExpiry,
	}, nil
}
