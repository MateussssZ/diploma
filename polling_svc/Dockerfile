FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
COPY api/ ./api/
COPY polling_svc/ ./polling_svc/
WORKDIR /workspace/polling_svc
RUN go mod download && go build -o /out/polling_svc ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/polling_svc .
COPY --from=builder /workspace/polling_svc/config.yaml .
EXPOSE 8084
CMD ["./polling_svc"]