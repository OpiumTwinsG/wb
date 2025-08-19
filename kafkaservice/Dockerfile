# syntax=docker/dockerfile:1

FROM golang:1.23-alpine as builder
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o kafkaService ./cmd
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /src/kafkaService ./
COPY .env ./
CMD  [ "./kafkaService" ]