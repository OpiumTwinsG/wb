package main

import (
	"context"
	"net/http"
	"orderservice/internal/cache"
	"orderservice/internal/config"
	"orderservice/internal/httpserver"
	"orderservice/internal/kafka"
	"orderservice/internal/pkg/logger"
	"orderservice/internal/repository"
	httphandlers "orderservice/internal/transport/httpHandlers"
	"orderservice/migrations"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	storage, err := repository.NewPostgresRepository(ctx, cfg.POSTGRESCONFIG)
	if err != nil {
		logger.Log.Fatalw("could not create DB", "err", err)
	}
	if err := migrations.CreateTables(ctx, storage.Pool); err != nil {
		logger.Log.Fatalw("could not create migrations", "err: ", err.Error())
	}
	cache := cache.NewCache(storage)

	handler := httphandlers.New(storage, cache)
	srv := httpserver.NewOrderServiceServer(":"+strconv.Itoa(cfg.HTTPPort), handler.NewRouter())

	//создание консюмера
	consumer, err := kafka.NewConsumer(
		cfg.KFKACONFIG.Kafkabrokers, cfg.KFKACONFIG.KfkagroupID, storage, cache)
	if err != nil {
		logger.Log.Fatalw("could not create consumer", "err:", err.Error())
	}
	go func() {
		if err := consumer.Run(ctx, cfg.KFKACONFIG.KfkaTopic); err != nil {
			logger.Log.Fatalw("could not run consumer service", "err:", err)
		}
	}()

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalw("http server error", "err", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // ждём Ctrl‑C или SIGTERM от Kubernetes

	srv.Stop() // корректно гасим HTTP‑сервер

}
