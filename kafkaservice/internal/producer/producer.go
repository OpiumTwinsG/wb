package producer

import (
	"encoding/json"
	"kafkaservice/internal/config"
	"kafkaservice/internal/pkg/logger"
	"time"
	"github.com/IBM/sarama"
)
type Producer struct{
	Async sarama.AsyncProducer
	Topic string
}

func NewAsyncProducer(cfg config.Config) *Producer{
	cfgSarama := sarama.NewConfig() //новый sarama Config
	cfgSarama.Producer.RequiredAcks = sarama.WaitForLocal //ждем от лидера подверждение
	cfgSarama.Producer.Retry.Max = 5 //число ретраев
	cfgSarama.Producer.Idempotent = true
	cfgSarama.Producer.Return.Successes = true
	cfgSarama.Producer.Return.Errors = true
	
	producer, err := sarama.NewAsyncProducer(cfg.KFKA_BROKERS, cfgSarama)
	if err !=nil{
		logger.Log.Fatalw("could not create asProd", "err", err)
	}
	defer producer.Close()
	logger.Log.Infow("create new producer")
	return &Producer{Async:producer, Topic: cfg.KFKA_TOPICS}
}

func (p *Producer) Send(evt interface{}){
	buf, err := json.Marshal(evt)
	if err != nil{
		logger.Log.Errorw("could not marshal data", "err", err)
	}
	msg := &sarama.ProducerMessage{
		Topic: p.Topic,
		Value: sarama.ByteEncoder(buf),
		Timestamp: time.Now(),
		
	}
	logger.Log.Infow("send msg to chan")
	p.Async.Input() <- msg
}

func (p *Producer) Close() error{
	logger.Log.Infow("close prod")
	return p.Async.Close()
}