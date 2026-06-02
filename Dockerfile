FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/analyzer ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/analyzer ./analyzer
COPY migrations ./migrations
COPY config ./config

EXPOSE 50051 9090 6060

ENTRYPOINT ["./analyzer"]
CMD ["serve", "--config", "config/docker.yaml", "--host", "0.0.0.0", "--port", "50051"]
