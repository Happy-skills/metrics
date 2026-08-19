package db

import (
	"context"
	"errors"
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
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", upErr)
	}

	return pgxPool, nil
}
