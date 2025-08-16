package migrations
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Создаем транзакцию для миграций
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаем таблицу orders
	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS orders(
			order_uid          VARCHAR PRIMARY KEY,
			track_number       VARCHAR NOT NULL,
			entry              VARCHAR NOT NULL,
			locale             VARCHAR NOT NULL,
			internal_signature VARCHAR,
			customer_id        VARCHAR NOT NULL,
			delivery_service   VARCHAR NOT NULL,
			shardkey           VARCHAR NOT NULL,
			sm_id              INTEGER NOT NULL,
			date_created       TIMESTAMP NOT NULL,
			oof_shard          VARCHAR NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	// Создаем таблицу deliveries
	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS deliveries(
			id         SERIAL PRIMARY KEY,
			order_uid  VARCHAR NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
			name       VARCHAR NOT NULL,
			phone      VARCHAR NOT NULL,
			zip        VARCHAR NOT NULL,
			city       VARCHAR NOT NULL,
			address    VARCHAR NOT NULL,
			region     VARCHAR NOT NULL,
			email      VARCHAR NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	// Создаем таблицу payments
	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS payments(
			transaction   VARCHAR PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
			request_id    VARCHAR,
			currency      VARCHAR(3) NOT NULL,
			provider      VARCHAR NOT NULL,
			amount        INTEGER NOT NULL,
			payment_dt    BIGINT NOT NULL,
			bank          VARCHAR NOT NULL,
			delivery_cost INTEGER NOT NULL,
			goods_total   INTEGER NOT NULL,
			custom_fee    INTEGER NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	// Создаем таблицу items
	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS items(
			id           SERIAL PRIMARY KEY,
			order_uid    VARCHAR NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
			chrt_id      INTEGER NOT NULL,
			track_number VARCHAR NOT NULL,
			price        INTEGER NOT NULL,
			rid          VARCHAR NOT NULL,
			name         VARCHAR NOT NULL,
			sale         INTEGER NOT NULL,
			size         VARCHAR NOT NULL,
			total_price  INTEGER NOT NULL,
			nm_id        INTEGER NOT NULL,
			brand        VARCHAR NOT NULL,
			status       INTEGER NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	if err := tx.Commit(ctx);err!=nil{
		return fmt.Errorf("failed to commit transaction:%w", err)
	}

	return nil
}