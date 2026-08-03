package repository

import (
	"context"
	"fmt"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(pool *pgxpool.Pool) Storage {
	return &dbStorage{pool: pool}
}

func (r *dbStorage) SetValue(ctx context.Context, mType string, mName string, mValue any) error {
	if mType == models.Gauge {
		_, err := r.pool.Exec(ctx, `insert into metric_values (name,type,value) values ($1,$2,$3) on conflict(name,type) do update set value = $3`, mName, mType, mValue)
		if err != nil {
			return fmt.Errorf("metric_values insert metric error: %w", err)
		}
	} else if mType == models.Counter {
		_, err := r.pool.Exec(ctx, `insert into metric_values (name,type,delta) values ($1,$2,$3) on conflict(name,type) do update set delta = metric_values.delta + $3`, mName, mType, mValue)
		if err != nil {
			return fmt.Errorf("metric_values insert metric error: %w", err)
		}
	}

	return nil
}

func (r *dbStorage) GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error) {
	metric := models.Metrics{}

	row := r.pool.QueryRow(ctx, `select name, type, value, delta from metric_values where name = $1 and type = $2`, mName, mType)

	if err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
		return nil, fmt.Errorf("metric_values scan metric error: %w", err)
	}

	return &metric, nil
}

func (r *dbStorage) GetValues(ctx context.Context) map[string]models.Metrics {
	ret := make(map[string]models.Metrics)

	rows, err := r.pool.Query(ctx, `select name, type, value, delta from metric_values`)
	if err != nil {
		logger.Sugar.Errorf("metric_values finding metric error: %w", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
			logger.Sugar.Errorf("metric_values finding metric error: %w", err)
			return nil
		}
		ret[metric.ID] = metric
	}

	return ret
}

func (r *dbStorage) ResetValue(mType string, mName string) error {
	//TODO implement me
	panic("implement me")
}

func (r *dbStorage) GetValuesSlice() []models.Metrics {
	//TODO implement me
	panic("implement me")
}

func (r *dbStorage) SetValuesFromSlice(metrics []models.Metrics) error {
	//TODO implement me
	panic("implement me")
}

func (r *dbStorage) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
