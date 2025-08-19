package config

import (
	
	"kafkaservice/internal/pkg/logger"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	KFKA_BROKERS []string `env:"KAFKA_BROKEKS"`
	KFKA_TOPICS  string `env:"KAFKA_TOPICS"`
	LOG_ENV string `env:"ENV"`
	LOG_LVL string `env:"LOG_LEVEL"`
}

func InitConfig() *Config{
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg);err!=nil{
		logger.InitLogger("error", "local")
		logger.Log.Fatalw("could not read env", "err", err)
	}
	logger.InitLogger(cfg.LOG_LVL, cfg.LOG_ENV)
	return cfg
}