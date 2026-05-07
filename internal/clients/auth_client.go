package clients

import (
	"context"

	authpb "apigateway/api/grpc/authservice"
	"apigateway/internal/api/rest/handlers/models"

	"google.golang.org/grpc"
)

// AuthServiceClient wraps the generated gRPC client and satisfies IUserCtrl.
type AuthServiceClient struct {
	client authpb.AuthServiceClient
}

func NewAuthServiceClient(conn *grpc.ClientConn) *AuthServiceClient {
	return &AuthServiceClient{client: authpb.NewAuthServiceClient(conn)}
}

func (c *AuthServiceClient) Register(ctx context.Context, req models.RegisterRequest) error {
	_, err := c.client.Register(ctx, &authpb.RegisterRequest{
		Login:    req.Login,
		Password: req.Password,
		Email:    req.Email,
	})
	return err
}

func (c *AuthServiceClient) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	resp, err := c.client.Login(ctx, &authpb.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &models.LoginResponse{
		AccessToken:        resp.AccessToken,
		RefreshToken:       resp.RefreshToken,
		AccessTokenExpiry:  resp.AccessTokenExpiry,
		RefreshTokenExpiry: resp.RefreshTokenExpiry,
	}, nil
}

func (c *AuthServiceClient) Logout(ctx context.Context, refreshToken string) error {
	_, err := c.client.Logout(ctx, &authpb.LogoutRequest{RefreshToken: refreshToken})
	return err
}

func (c *AuthServiceClient) Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error) {
	resp, err := c.client.Refresh(ctx, &authpb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, err
	}
	return &models.RefreshResponse{
		AccessToken:        resp.AccessToken,
		RefreshToken:       resp.RefreshToken,
		AccessTokenExpiry:  resp.AccessTokenExpiry,
		RefreshTokenExpiry: resp.RefreshTokenExpiry,
	}, nil
}
