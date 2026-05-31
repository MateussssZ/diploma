FROM golang:1.25-alpine AS builder

# ca-certificates нужны для TLS-вызовов к внешним сервисам (gRPC/Kafka)
RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/apigateway ./cmd

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/apigateway /apigateway

# Конфиги монтируются через docker-compose volume, но кладём дефолты как fallback
COPY --from=builder /src/config.yaml /config.yaml
COPY --from=builder /src/credentials.yaml /credentials.yaml

EXPOSE 8080 9090

ENTRYPOINT ["/apigateway"]
