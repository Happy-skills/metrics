package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	PingDB(ctx context.Context) error
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) PingDB(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
