package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	var pgxPool *pgxpool.Pool

	if dsn != "" {
		connectionConfig, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to postgres connection config: %w", err)
		}

		pgxPool, err = pgxpool.NewWithConfig(ctx, connectionConfig)
		if err != nil {
			return nil, fmt.Errorf("database connection failed: %w", err)
		}

		if err := pgxPool.Ping(ctx); err != nil {
			return nil, fmt.Errorf("database ping failed: %w", err)
		}
	}

	return pgxPool, nil
}
