package grpc

// apigatewayService is the gRPC service stub.
// Register specific service implementations via server.go when protos are ready.
type apigatewayService struct{}

func NewApigatewayService() (*apigatewayService, error) {
	return &apigatewayService{}, nil
}
