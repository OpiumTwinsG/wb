package kafka

import (
  "context"
  "encoding/json"
  "github.com/IBM/sarama"
  "orderservice/internal/model"
  "orderservice/internal/cache"
//   "orderservice/internal/transport/httphandlers"
)

type ConfigKafka struct{
  Kafkabrokers []string `env:"KAFKA_BROKEKS"`
  KfkagroupID string `env:"KAFKA_GROUP_ID"`
  KfkaTopic []string `env:"KAFKA_TOPICS"`
}

type consumerHandler struct {
  ready chan struct{} // Канал для сигнализации о готовности consumer
  uc     OrderService// Слой бизнес-логики
  cache *cache.Cache
}

type Consumer struct {
  group   sarama.ConsumerGroup //// Группа потребителей Kafka
  handler *consumerHandler //// Обработчик сообщений
}
type OrderService interface{
  SaveOrder(ctx context.Context, order *model.Order) error
  GetOrders(ctx context.Context) ([]model.Order,error)
  GetOrderByID(ctx context.Context, id string) (model.Order,error)
}
//brokers - адреса брокеров кафки например "kafka1:9092", "kafka2:9092"
func NewConsumer(brokers []string, groupID string, uc OrderService, cache *cache.Cache) (*Consumer, error) {
  cfg := sarama.NewConfig()
  cfg.Version = sarama.V2_8_0_0 // Указываем версию Kafka
  cfg.Consumer.Offsets.Initial = sarama.OffsetNewest // Читать с новых сообщений
  // Создаем группу потребителей
  g, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
  if err != nil {
    return nil, err
  }
  // Инициализируем обработчик
  h := &consumerHandler{ready: make(chan struct{}), uc: uc, cache: cache}
  return &Consumer{group: g, handler: h}, nil
}
//Setup - вызывается при инициализации
func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
  close(h.ready) //// Сигнализируем, что consumer готов к работе
  return nil
}
// Cleanup - вызывается при завершении
func (h *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
  return nil// Здесь можно добавить логику очистки
}
//ConsumeClaim - обработка сообщений
func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
   // Читаем сообщения из партиции
  for msg := range claim.Messages() {
    var order model.Order
  // Декодируем JSON сообщение
    if err := json.Unmarshal(msg.Value, &order); err != nil {
      continue
    }
   // Сохраняем заказ через use case
    if err := h.uc.SaveOrder(sess.Context(), &order); err != nil {
      continue
    }
  h.cache.Set(order.OrderUID, order)
  // Подтверждаем обработку сообщения
    sess.MarkMessage(msg, "")
  }

  return nil
}
// Run - запуск consumer
func (c *Consumer) Run(ctx context.Context, topics []string)error {
   errChan := make(chan error, 1) // Буферизированный канал
    
    go func() {
        for {
            if err := c.group.Consume(ctx, topics, c.handler); err != nil {
                errChan <- err // Отправляем ошибку в канал
                return
            }
            if ctx.Err() != nil {
                return
            }
        }
    }()
    
    select {
    case <-c.handler.ready:
        return nil // Успешный старт
    case err := <-errChan:
        return err // Ошибка при запуске
    case <-ctx.Done():
        return ctx.Err() // Отмена контекста
    }
}
// Close - завершение работы
func (c *Consumer) Close() error {
  return c.group.Close() //// Корректное закрытие consumer
}