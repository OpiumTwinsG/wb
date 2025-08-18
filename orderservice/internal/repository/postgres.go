package repository

import (
	"context"
	// "encoding/json"
	"fmt"
	// "time"
	"orderservice/internal/model"
	"orderservice/internal/pkg/logger"

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
	Pool *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, cfg Config) (*Storage , error){
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.UserName, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	Pool, err := pgxpool.New(ctx, dsn)
	if err !=nil{
        logger.Log.Errorw("failed to connect to pool","err", err)
		return nil, fmt.Errorf("connect to db: %w", err)
	}
    logger.Log.Infow("get connection to pool", "dsn: ", dsn)
	err = Pool.Ping(ctx)
	if err !=nil{
        logger.Log.Errorw("failed go get connection to pool", "err", err)
		return nil, fmt.Errorf("ping to db: %w",err)
	}
    logger.Log.Infow("successfully connect to db")
	return &Storage{Pool: Pool}, nil
}

func (s *Storage) Close() {
    logger.Log.Debugw("Close pool")
    s.Pool.Close()
}

func (s *Storage) SaveOrder(ctx context.Context, order *model.Order) error{
	tx, err := s.Pool.Begin(ctx)
	 if err != nil {
        logger.Log.Errorw("failed to begin transaction", "err", err)
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    logger.Log.Debugw("successfully begin transaction")
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
        logger.Log.Errorw("failed to insert order", "err", err, "order_uid", order.OrderUID,)
        return fmt.Errorf("failed to insert order: %w", err)
    }
    logger.Log.Debugw("Successfully insert into orders", "order_uid", order.OrderUID, "customer_id", order.CustomerID)
	const deliveryQuery = `INSERT INTO deliveries (
        order_uid, name, phone, zip, city, address, region, email
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(ctx, deliveryQuery,
        order.OrderUID, order.Delivery.Name, order.Delivery.Phone, 
        order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
        order.Delivery.Region, order.Delivery.Email)
    if err != nil {
        logger.Log.Errorw("failed to insert delivery", "err", err, "order_uid", order.OrderUID, "delivery_name", order.Delivery.Name)
        return fmt.Errorf("failed to insert delivery: %w", err)
    }
    logger.Log.Debugw("Successfully insert into deliveries", "order_uid", order.OrderUID, "customer_id", order.CustomerID,"delivery_name", order.Delivery.Name)
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
        logger.Log.Errorw("faild to insert payment", "err", err, "order_uid", order.OrderUID, "transaction", order.Payment.Transaction)
        return fmt.Errorf("failed to insert payment: %w", err)
    }
    logger.Log.Debugw("Successfully insert into payments", "order_uid", order.OrderUID, "customer_id", order.CustomerID, "transaction", order.Payment.Transaction)
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
            logger.Log.Errorw("failed to insert item", "err", err, "order_uid", order.OrderUID)
            return fmt.Errorf("failed to insert item: %w", err)
        }
	}
    logger.Log.Debugw("Successfully insert into items", "order_uid", order.OrderUID)
	if err := tx.Commit(ctx); err!=nil{
        logger.Log.Errorw("failed to commit transaction", "err",err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
    logger.Log.Infow("successfulle commit transaction")
	return nil
}

func (s *Storage) GetOrders(ctx context.Context) ([]model.Order,error){
	var orders []model.Order
	const q = `select order_uid from orders`
	rows, err := s.Pool.Query(ctx, q)
	if err !=nil{
		return nil, err
	}
    logger.Log.Debugw("successfulle get rows")
	defer rows.Close()
	for rows.Next() {
        var orderUID string
        if err := rows.Scan(&orderUID);err!=nil{
			return nil, fmt.Errorf("failed to scan order uid: %w", err)
		}
        logger.Log.Debugw("successfulle scan into orderUID")
		order, err := s.GetOrderByID(ctx, orderUID)
		if err !=nil{
			return nil, fmt.Errorf("failed to load full order %s: %w",orderUID, err)
		}
        logger.Log.Debugw("successfulle GetOrderByID")
        orders = append(orders, order)
    }
    logger.Log.Infow("successfully return orders")
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
	_ = s.Pool.QueryRow(ctx, orderQuery, id)
	err := s.Pool.QueryRow(ctx, orderQuery, id).Scan(
        &order.OrderUID, &order.TrackNumber, &order.Entry,
        &order.Locale, &order.InternalSignature, &order.CustomerID,
        &order.DeliveryService, &order.Shardkey, &order.SmID,
        &order.DateCreated, &order.OofShard)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get order: %w", err)
    }
    logger.Log.Debugw("successfully get order")
	// 2. Получаем доставку
    const deliveryQuery = `SELECT 
        name, phone, zip, city, address, region, email
    FROM deliveries WHERE order_uid = $1`
    
    order.Delivery = &model.Delivery{OrderUID: id}
    err = s.Pool.QueryRow(ctx, deliveryQuery, id).Scan(
        &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
        &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
        &order.Delivery.Email)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get delivery: %w", err)
    }
    logger.Log.Debugw("successfully get delivery")
	// 3. Получаем платеж
    const paymentQuery = `SELECT 
        request_id, currency, provider, amount,
        payment_dt, bank, delivery_cost, goods_total, custom_fee
    FROM payments WHERE transaction = $1`
    
    order.Payment = &model.Payment{Transaction: id}
    err = s.Pool.QueryRow(ctx, paymentQuery, id).Scan(
        &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
        &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
        &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to get payment: %w", err)
    }
    logger.Log.Debugw("successfully get payment")
	// 4. Получаем товары
    const itemsQuery = `SELECT 
        chrt_id, track_number, price, rid, name,
        sale, size, total_price, nm_id, brand, status
    FROM items WHERE order_uid = $1`
	rows, err := s.Pool.Query(ctx, itemsQuery, id)
	if err !=nil{
		return order, fmt.Errorf("failed to get rows by id %s: %w",id, err)
	}
    logger.Log.Debugw("successfully get items")
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
        logger.Log.Debugw("successfully get items")
        order.Items = append(order.Items, item)
	}
    logger.Log.Infow("successfully get order")
	return order, nil
}