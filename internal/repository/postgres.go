package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
	mu   sync.Mutex
}

func NewDBStorage(pool *pgxpool.Pool) Storage {
	return &dbStorage{pool: pool}
}

func (r *dbStorage) SetValue(ctx context.Context, mType string, mName string, mValue any) error {
	var err error
	if mType == models.Gauge {
		err = r.execute(
			ctx,
			`insert into metric_values (name,type,value) values ($1,$2,$3) on conflict(name,type) do update set value = $3`,
			mName, mType, mValue,
		)

	} else if mType == models.Counter {
		err = r.execute(
			ctx,
			`insert into metric_values (name,type,delta) values ($1,$2,$3) on conflict(name,type) do update set delta = metric_values.delta + $3`,
			mName, mType, mValue,
		)
	}
	if err != nil {
		return fmt.Errorf("metric_values insert metric error: %w", err)
	}

	return nil
}

func (r *dbStorage) GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error) {
	metric := models.Metrics{}

	rows, err := r.query(ctx, `select name, type, value, delta from metric_values where name = $1 and type = $2 limit 1`, mName, mType)
	if err != nil {
		logger.Sugar.Errorf("metric_values finding metric error: %s", err.Error())
		return nil, fmt.Errorf("metric_values finding metric error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
			logger.Sugar.Errorf("metric_values scan metric error: %s", err.Error())
			return nil, fmt.Errorf("metric_values scan metric error: %w", err)
		}
	}

	if metric.ID == "" {
		return nil, fmt.Errorf("metric_values metric not found")
	}

	return &metric, nil
}

func (r *dbStorage) GetValues(ctx context.Context) map[string]models.Metrics {
	ret := make(map[string]models.Metrics)

	rows, err := r.query(ctx, `select name, type, value, delta from metric_values`)
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

func (r *dbStorage) Begin(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	r.tx = tx

	return nil
}

func (r *dbStorage) Commit(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tx == nil {
		return fmt.Errorf("transaction not started")
	}

	if err := r.tx.Commit(ctx); err != nil {
		return err
	}

	r.tx = nil

	return nil
}

func (r *dbStorage) Rollback(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tx == nil {
		return fmt.Errorf("transaction not started")
	}

	if err := r.tx.Rollback(ctx); err != nil {
		return err
	}

	r.tx = nil

	return nil
}

func (r *dbStorage) execute(ctx context.Context, query string, arguments ...any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tx != nil {
		_, err := r.tx.Exec(ctx, query, arguments...)
		return err
	}

	_, err := r.pool.Exec(ctx, query, arguments...)
	return err
}

func (r *dbStorage) query(ctx context.Context, query string, arguments ...any) (pgx.Rows, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tx != nil {
		return r.tx.Query(ctx, query, arguments...)
	}

	return r.pool.Query(ctx, query, arguments...)
}
