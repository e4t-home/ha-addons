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

RUN apk add --no-cache bash

COPY --from=builder /app/gpsd-proxy /usr/local/bin/
COPY run.sh /
RUN chmod a+x /run.sh

CMD [ "/run.sh" ]
