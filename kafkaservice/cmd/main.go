package main

import (
	"kafkaservice/internal/config"
	"kafkaservice/internal/pkg/logger"
	"kafkaservice/internal/producer"
	"time"
)

func main() {
	cfg := config.InitConfig()
	logger.InitLogger(cfg.LOG_LVL, cfg.LOG_ENV)
	prod := producer.NewAsyncProducer(*cfg)
	defer prod.Close()

	go func(){
		for msg := range prod.Async.Successes(){
			logger.Log.Infof("successfully push message to 	topic=%s partition=%d offset=%d", msg.Topic, msg.Partition, msg.Offset)
		}
	}()
	go func(){
		for error := range prod.Async.Errors(){
			logger.Log.Errorw("error prod", "err",error)
		}
	}()
	go func(){
		for i:=0; i < 1000;i++ {
			prod.Send()


				
			time.Sleep(1 * time.Second)
		}
	}()
}