package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustPool(dsn string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil { panic(err) }
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil { panic(err) }
	return pool
}

