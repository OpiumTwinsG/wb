package main

import (
	"context"
	"orderservice/internal/config"
	"orderservice/internal/repository"
	"orderservice/internal/cache"
	"orderservice/internal/transport/httpHandlers"
	"orderservice/internal/httpserver"
	"orderservice/internal/pkg/logger"
	"strconv"
	"net/http"
	"syscall"
	"os/signal"
	"os"
)

func main(){
	ctx := context.Background()
	cfg := config.LoadConfig()
	storage, err := repository.NewPostgresRepository(ctx, cfg.POSTGRESCONFIG)
	if err !=nil{
		logger.Log.Fatalw("could not create DB", "err", err)
	}
	cache := cache.NewCache(storage)
	handler := httphandlers.New(storage, cache)
	srv := httpserver.NewOrderServiceServer(":"+strconv.Itoa(cfg.HTTPPort), handler.NewRouter())
	go func(){
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalw("http server error", "err", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // ждём Ctrl‑C или SIGTERM от Kubernetes

	srv.Stop() // корректно гасим HTTP‑сервер
	
}