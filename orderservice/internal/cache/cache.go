package cache

import (
	model "orderservice/internal/model"
	"orderservice/internal/pkg/logger"
	"sync"

	// "orderservice/internal/repository"
	"context"
	"fmt"
)
type OrderService interface{
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrders(ctx context.Context) ([]model.Order,error)
	GetOrderByID(ctx context.Context, id string) (model.Order,error)
}
type Cache struct {
	cache map[string]model.Order
	mu    sync.RWMutex
	Repo OrderService  
}

func NewCache( repo OrderService) *Cache {
	cache := &Cache{cache: make(map[string]model.Order), Repo: repo}
	if err := cache.restoreFromDB();err!=nil{
		logger.Log.Errorf("could not load data from database %v", err)
	}
	return cache
}

func (c *Cache) restoreFromDB() error{

	c.mu.Lock()
	defer c.mu.Unlock()

	orders, err := c.Repo.GetOrders(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get orders from DB: %w", err)
	}
	for _, order := range orders{
		c.cache[order.OrderUID] = order
	}
	logger.Log.Infof("Restored %d orders from DB to cache", len(orders))

	return nil
}

func (c *Cache) Set(k string, v model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[k] = v
	c.cache[k]=v
	logger.Log.Debugw("successfully set order by key:", "key:",k)
	
}
func (c *Cache) Get(k string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.cache[k]
	logger.Log.Debugw("successfully get order by key", "key", k, "found", ok)
	return order, ok
}

func (c *Cache) GetAll() map[string]model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copy := make(map[string]model.Order, len(c.cache))
	for k, v := range c.cache{
		copy[k] = v
	}
	logger.Log.Debugw("sucessfully receive all elements")
	return copy
}