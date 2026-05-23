// client.go — gRPC client wrappers for polling_svc.
// Uses the same generated proto packages as the main apigateway module.
package client

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	auctionpb "apigateway/api/grpc/auctionservice"
	authpb "apigateway/api/grpc/authservice"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const grpcTimeout = 4 * time.Second

// AuctionClient wraps the two AuctionService gRPC stubs.
type AuctionClient struct {
	lots auctionpb.LotGrpcServiceClient
	bids auctionpb.BidGrpcServiceClient
}

func NewAuctionClient(address string) (*AuctionClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial AuctionService: %w", err)
	}
	return &AuctionClient{
		lots: auctionpb.NewLotGrpcServiceClient(conn),
		bids: auctionpb.NewBidGrpcServiceClient(conn),
	}, conn, nil
}

// AuthClient wraps the AuthService gRPC stub (unix socket).
type AuthClient struct {
	stub authpb.AuthServiceClient
}

func NewAuthClient(socketPath string) (*AuthClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"passthrough:///authservice",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial AuthService: %w", err)
	}
	return &AuthClient{stub: authpb.NewAuthServiceClient(conn)}, conn, nil
}

// ─── AuctionClient methods ────────────────────────────────────────────────────

func (c *AuctionClient) GetLots(ctx context.Context, page, size int32) (*auctionpb.GetLotsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.lots.GetLots(ctx, &auctionpb.GetLotsRequest{Page: page, Size: size})
}

func (c *AuctionClient) GetLot(ctx context.Context, id int64) (*auctionpb.LotFull, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.lots.GetLot(ctx, &auctionpb.GetLotRequest{Id: id})
}

func (c *AuctionClient) CreateLot(ctx context.Context, req *auctionpb.CreateLotRequest) (*auctionpb.LotFull, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.lots.CreateLot(ctx, req)
}

func (c *AuctionClient) PlaceBid(ctx context.Context, lotID, bidderID int64, amount string) (*auctionpb.BidProto, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.bids.PlaceBid(ctx, &auctionpb.PlaceBidRequest{
		LotId:    lotID,
		BidderId: bidderID,
		Amount:   amount,
	})
}

// ─── AuthClient methods ───────────────────────────────────────────────────────

func (c *AuthClient) Register(ctx context.Context, login, password, email string) error {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.stub.Register(ctx, &authpb.RegisterRequest{
		Login: login, Password: password, Email: email,
	})
	return err
}

func (c *AuthClient) Login(ctx context.Context, login, password string) (*authpb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.stub.Login(ctx, &authpb.LoginRequest{Login: login, Password: password})
}

func (c *AuthClient) Logout(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.stub.Logout(ctx, &authpb.LogoutRequest{RefreshToken: refreshToken})
	return err
}

func (c *AuthClient) Refresh(ctx context.Context, refreshToken string) (*authpb.RefreshResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	return c.stub.Refresh(ctx, &authpb.RefreshRequest{RefreshToken: refreshToken})
}

// ─── Price helper ─────────────────────────────────────────────────────────────

func ParsePrice(s string) int64 {
	f, _ := strconv.ParseFloat(s, 64)
	return int64(f * 100)
}
