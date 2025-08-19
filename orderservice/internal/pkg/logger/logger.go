package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func Init(level, env string){
	var cfg zap.Config
	if env == "local" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	l, err := cfg.Build()
	if err != nil {
		panic("cannot init zap logger: " + err.Error())
	}
	Log = l.Sugar()
}