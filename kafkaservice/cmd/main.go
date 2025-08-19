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
	order := []byte(`{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ]
}`)
	for {
        prod.SendBytes(order)
        time.Sleep(1 * time.Second)
    }
}