package repository

import (
	"context"
	"fmt"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/retrier"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertGaugeSQL   = `insert into metric_values (name,type,value) values ($1,$2,$3) on conflict(name,type) do update set value = $3`
	insertCounterSQL = `insert into metric_values (name,type,delta) values ($1,$2,$3) on conflict(name,type) do update set delta = metric_values.delta + $3`
	selectValueSQL   = `select name, type, value, delta from metric_values where name = $1 and type = $2 limit 1`
	selectValuesSQL  = `select name, type, value, delta from metric_values`
)

type dbStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(pool *pgxpool.Pool) Storage {
	return &dbStorage{pool: pool}
}

func (r *dbStorage) SetValue(ctx context.Context, mType string, mName string, mValue any) error {
	switch mType {
	case models.Gauge:
		value, err := toGauge(mValue)
		if err != nil {
			return err
		}

		if err := r.executeWithRetry(ctx, r.pool, insertGaugeSQL, mName, models.Gauge, value); err != nil {
			return fmt.Errorf("unable to insert gauge: %w", err)
		}
	case models.Counter:
		value, err := toCounter(mValue)
		if err != nil {
			return err
		}

		if err := r.executeWithRetry(ctx, r.pool, insertCounterSQL, mName, models.Counter, value); err != nil {
			return fmt.Errorf("unable to insert counter: %w", err)
		}
	}

	return nil
}

func (r *dbStorage) GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error) {
	metric := models.Metrics{}

	rows, err := r.query(ctx, r.pool, selectValueSQL, mName, mType)
	if err != nil {
		return nil, fmt.Errorf("metric_values finding metric error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
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

	rows, err := r.query(ctx, r.pool, selectValuesSQL)
	if err != nil {
		logger.Sugar.Errorf("query values: %s", err.Error())
		return ret
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics

		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
			logger.Sugar.Errorf("scan values: %s", err.Error())
			return ret
		}

		ret[metric.ID] = metric
	}

	return ret
}

func (r *dbStorage) SetValues(ctx context.Context, metrics []models.Metrics) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			batch.Queue(insertGaugeSQL, m.ID, models.Gauge, *m.Value)
		case models.Counter:
			batch.Queue(insertCounterSQL, m.ID, models.Counter, *m.Delta)
		}
	}

	if err := r.executeBatchWithRetry(ctx, tx, batch); err != nil {
		return fmt.Errorf("execute batch error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction error: %w", err)
	}

	return nil
}

func (r *dbStorage) ResetValue(ctx context.Context, mType string, mName string) error {
	return r.SetValue(ctx, mType, mName, 0)
}

func (r *dbStorage) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

type db interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func (r *dbStorage) query(ctx context.Context, db db, query string, arguments ...any) (pgx.Rows, error) {
	return db.Query(ctx, query, arguments...)
}

func (r *dbStorage) executeBatchWithRetry(ctx context.Context, db db, batch *pgx.Batch) error {
	return retrier.Retrier(func() (isRetriable bool, err error) {
		err = db.SendBatch(ctx, batch).Close()
		if err == nil {
			return false, nil
		}

		return classify(err) == Retriable,
			fmt.Errorf("error batching queries: %w", err)
	})
}

func (r *dbStorage) executeWithRetry(ctx context.Context, db db, query string, arguments ...any) error {
	return retrier.Retrier(func() (isRetriable bool, err error) {
		_, err = db.Exec(ctx, query, arguments...)
		if err == nil {
			return false, nil
		}

		return classify(err) == Retriable,
			fmt.Errorf("error executing query %s: %w", query, err)
	})
}
