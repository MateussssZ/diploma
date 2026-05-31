FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o auth_service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/auth_service .
COPY migrations/ ./migrations/
COPY config.yaml .
EXPOSE 8083
CMD ["./auth_service"]
