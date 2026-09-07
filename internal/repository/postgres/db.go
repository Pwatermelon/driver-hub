package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) error {
	data, err := os.ReadFile(migrationsPath)
	if err != nil {
		return err
	}
	sql := string(data)
	// Простой runner: берём только Up-часть до Down.
	if i := strings.Index(sql, "-- +migrate Down"); i >= 0 {
		sql = sql[:i]
	}
	_, err = pool.Exec(ctx, sql)
	return err
}

func WaitReady(ctx context.Context, databaseURL string, attempts int) (*pgxpool.Pool, error) {
	var last error
	for i := 0; i < attempts; i++ {
		pool, err := Connect(ctx, databaseURL)
		if err == nil {
			return pool, nil
		}
		last = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("db not ready: %w", last)
}
