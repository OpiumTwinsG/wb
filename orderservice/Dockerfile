# syntax=docker/dockerfile:1

# ── Build Stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
WORKDIR /src
# 1) Копируем только манифест для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# 2) Копируем весь исходный код сервиса
COPY . .

# 3) Собираем статический бинарь
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o dbService ./cmd

# ── Runtime Stage ────────────────────────────────────────────────────────────
FROM alpine:3.19
WORKDIR /app

# Копируем бинарь из билд-стадии
COPY --from=builder /src/orderService ./

# Копируем файл настроек .env внутрь контейнера
COPY .env ./

# Открываем порт, на котором слушает gRPC
# EXPOSE 50051
EXPOSE 8080
# Запуск приложения
ENTRYPOINT ["./orderService"]