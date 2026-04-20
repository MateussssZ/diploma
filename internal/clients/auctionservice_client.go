package clients

import (
	auctionpb "apigateway/api/grpc/auctionservice"

	"google.golang.org/grpc"
)

type AuctionServiceClient struct {
	LotClient auctionpb.LotGrpcServiceClient
	BidClient auctionpb.BidGrpcServiceClient
}

func NewAuctionServiceClient(conn *grpc.ClientConn) *AuctionServiceClient {
	return &AuctionServiceClient{
		LotClient: auctionpb.NewLotGrpcServiceClient(conn),
		BidClient: auctionpb.NewBidGrpcServiceClient(conn),
	}
}
