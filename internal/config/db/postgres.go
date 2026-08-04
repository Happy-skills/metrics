package db

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	return pgxPool, nil
}
