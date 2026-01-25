ARG BUILD_FROM
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o matter-code-db .

FROM $BUILD_FROM

RUN apk add --no-cache sqlite-libs

WORKDIR /app
COPY --from=builder /app/matter-code-db .
COPY templates/ ./templates/
COPY static/ ./static/

COPY run.sh /
RUN chmod a+x /run.sh

CMD [ "/run.sh" ]
