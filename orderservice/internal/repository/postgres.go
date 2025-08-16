package repository

import (
	"context"
	// "encoding/json"
	"fmt"
	// "time"
	"orderservice/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	UserName string `env:"POSTGRES_USER"     env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `env:"POSTGRES_HOST"     env-required:"true"`
	Port     string `env:"POSTGRES_PORT"     env-required:"true"`
	DBName   string `env:"POSTGRES_DB"       env-required:"true"`
}

type Storage   struct{
	pool *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, cfg Config) (*Storage , error){
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.UserName, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err !=nil{
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	err = pool.Ping(ctx)
	if err !=nil{
		return nil, fmt.Errorf("ping to db: %w",err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {s.pool.Close()}

func (s *Storage) SaveOrder(ctx context.Context, order *model.Order) error{
	tx, err := s.pool.Begin(ctx)
	 if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)
	 // 1. Сохраняем основной заказ
    const orderQuery = `INSERT INTO orders (
        order_uid, track_number, entry, locale, 
        internal_signature, customer_id, delivery_service,
        shardkey, sm_id, date_created, oof_shard
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(ctx, orderQuery,
        order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
        order.InternalSignature, order.CustomerID, order.DeliveryService,
        order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
    if err != nil {
        return fmt.Errorf("failed to insert order: %w", err)
    }
	const deliveryQuery = `INSERT INTO deliveries (
        order_uid, name, phone, zip, city, address, region, email
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(ctx, deliveryQuery,
        order.OrderUID, order.Delivery.Name, order.Delivery.Phone, 
        order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
        order.Delivery.Region, order.Delivery.Email)
    if err != nil {
        return fmt.Errorf("failed to insert delivery: %w", err)
    }
	// 3. Сохраняем платеж
    const paymentQuery = `INSERT INTO payments (
        transaction, request_id, currency, provider, amount,
        payment_dt, bank, delivery_cost, goods_total, custom_fee
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
    
    _, err = tx.Exec(ctx, paymentQuery,
        order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
        order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt,
        order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
        order.Payment.CustomFee)
    if err != nil {
        return fmt.Errorf("failed to insert payment: %w", err)
    }
	// 4. Сохраняем товары
    const itemQuery = `INSERT INTO items (
        order_uid, chrt_id, track_number, price, rid, 
        name, sale, size, total_price, nm_id, brand, status
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items{
		_, err = tx.Exec(ctx, itemQuery,
            order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
            item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
            item.Brand, item.Status)
        if err != nil {
            return fmt.Errorf("failed to insert item: %w", err)
        }
	}
	if err := tx.Commit(ctx); err!=nil{
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *Storage) GetOrders(ctx context.Context) ([]model.Order,error){
	var orders []model.Order
	const q = `select order_uid from orders`
	rows, err := s.pool.Query(ctx, q)
	if err !=nil{
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
        var orderUID string
        if err := rows.Scan(&orderUID);err!=nil{
			return nil, fmt.Errorf("failed to scan order uid: %w", err)
		}
		order, err := s.GetOrderByID(ctx, orderUID)
		if err !=nil{
			return nil, fmt.Errorf("failed to load full order %s: %w",orderUID, err)
		}
        
        orders = append(orders, order)
    }
	return orders, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id string) (model.Order,error){
	var order model.Order
	// 1. Получаем основной заказ
    const orderQuery = `SELECT 
        order_uid, track_number, entry, locale, 
        internal_signature, customer_id, delivery_service,
        shardkey, sm_id, date_created, oof_shard
    FROM orders WHERE order_uid = $1`
	_ = s.pool.QueryRow(ctx, orderQuery, id)
	err := s.pool.QueryRow(ctx, orderQuery, id).Scan(
        &order.OrderUID, &order.TrackNumber, &order.Entry,
        &order.Locale, &order.InternalSignature, &order.CustomerID,
        &order.DeliveryService, &order.Shardkey, &order.SmID,
        &order.DateCreated, &order.OofShard)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get order: %w", err)
    }
	// 2. Получаем доставку
    const deliveryQuery = `SELECT 
        name, phone, zip, city, address, region, email
    FROM deliveries WHERE order_uid = $1`
    
    order.Delivery = &model.Delivery{OrderUID: id}
    err = s.pool.QueryRow(ctx, deliveryQuery, id).Scan(
        &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
        &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
        &order.Delivery.Email)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get delivery: %w", err)
    }
	// 3. Получаем платеж
    const paymentQuery = `SELECT 
        request_id, currency, provider, amount,
        payment_dt, bank, delivery_cost, goods_total, custom_fee
    FROM payments WHERE transaction = $1`
    
    order.Payment = &model.Payment{Transaction: id}
    err = s.pool.QueryRow(ctx, paymentQuery, id).Scan(
        &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
        &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
        &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get payment: %w", err)
    }
	// 4. Получаем товары
    const itemsQuery = `SELECT 
        chrt_id, track_number, price, rid, name,
        sale, size, total_price, nm_id, brand, status
    FROM items WHERE order_uid = $1`
	rows, err := s.pool.Query(ctx, itemsQuery, id)
	if err !=nil{
		return order, fmt.Errorf("failed to get rows by id %s: %w",id, err)
	}
	defer rows.Close()
	for rows.Next(){
		var item model.Item
        item.OrderUID = id
        err := rows.Scan(
            &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
            &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
        if err != nil {
            return model.Order{}, fmt.Errorf("failed to scan item: %w", err)
        }
        order.Items = append(order.Items, item)
	}

	return order, nil
}