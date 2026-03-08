# Build stage
ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.18
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o gpsd-proxy .

# Runtime stage
FROM ${BUILD_FROM}

COPY --from=builder /app/gpsd-proxy /usr/local/bin/

# Set up s6 service
COPY run.sh /etc/services.d/gpsd-proxy/run
RUN chmod a+x /etc/services.d/gpsd-proxy/run
