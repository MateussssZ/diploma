//go:generate protoc --proto_path=proto --go_out=api/auth --go_opt=paths=source_relative --go-grpc_out=api/auth --go-grpc_opt=paths=source_relative proto/auth.proto

package api
