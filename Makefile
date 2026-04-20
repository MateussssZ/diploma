.PHONY: help build lint start proto-install-tools proto-gen

MAKEFLAGS   += --no-print-directory
GIT_BRANCH  := $(shell git branch --show-current)
PROTO_DIR   := /api
PROTO_FILES := $(wildcard .$(PROTO_DIR)/*.proto)
MODULE_NAME := $(shell go list -m)
GO_OUT_DIR  := $(PROTO_DIR)/grpc
DOC_OUT_DIR := /docs

help:
	@echo "build                - Build for Linux"
	@echo "lint                 - Run golangci-lint"
	@echo "start                - Run the application locally"
	@echo "proto-install-tools  - Install protoc tools"
	@echo "proto-gen            - Generating protobuf code"


build:
	@echo "[INFO] Branch: $(GIT_BRANCH) - Building for Linux..."
	@set GOOS=linux&& set CGO_ENABLED=0&& go build -o portal ./cmd/main.go
	@echo "[OK] Linux build completed!"


lint:
	@echo "[INFO] Running golangci-lint..."
	@go mod tidy
	@golangci-lint run ./...
	@echo "[OK] Lint completed successfully!"


start:
	@echo "[INFO] Starting the application..."
	@go run ./cmd/main.go


proto-install-tools:
	@echo "[INFO] Installing protoc-tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	@echo "[OK] Protoc-tools installed successfully!"


proto-gen:
	@echo "[INFO] Generating protobuf code and docs..."
	@protoc --version
	@protoc \
		--proto_path=. \
		--go_out=.$(GO_OUT_DIR) \
		--go_opt=module=$(MODULE_NAME)$(GO_OUT_DIR) \
		--go-grpc_out=.$(GO_OUT_DIR) \
		--go-grpc_opt=module=$(MODULE_NAME)$(GO_OUT_DIR) \
		--doc_out=.$(DOC_OUT_DIR) \
		--doc_opt=markdown,api.md \
		$(PROTO_FILES)
	@echo "[OK] Protobuf code and docs generated successfully!"
