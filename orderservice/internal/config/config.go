package config

import (
	"awesomeproject7/orderservice/internal/repository"
	"github.com/ilyakaznacheev/cleanenv"
	"awesomeproject7/orderservice/internal/pkg/logger"
	"awesomeproject7/orderservice/internal/kafka"
)

type Config struct{
	POSTGRESCONFIG repository.Config
	HTTPPort int  `env:"HTTP_PORT" env-default:"8080"`
	KFKACONFIG kafka.ConfigKafka 
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	Env      string `env:"ENV" env-default:"local"`
}// LoadConfig читает .env / OS env и инициализирует zap‑логер.
func LoadConfig() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		logger.Init("error", "local")
		logger.Log.Fatalw("config read error", "err", err)
	}
	logger.Init(cfg.LogLevel, cfg.Env)
	return cfg
}