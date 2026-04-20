//go:build ignore

package api

//go:generate protoc --proto_path=./.. --go_out=./../api/grpc --go_opt=module=apigateway/api/grpc --go-grpc_out=./../api/grpc --go-grpc_opt=module=apigateway/api/grpc --doc_out=./../docs --doc_opt=markdown,api.md ./../api/template_messages.proto ./../api/template_service.proto
