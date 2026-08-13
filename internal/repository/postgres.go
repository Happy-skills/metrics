package repository

import (
	"context"
	"fmt"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

		if err := r.insertGauge(ctx, r.pool, mName, value); err != nil {
			return fmt.Errorf("unable to insert gauge: %w", err)
		}
	case models.Counter:
		value, err := toCounter(mValue)
		if err != nil {
			return err
		}

		if err := r.insertCounter(ctx, r.pool, mName, value); err != nil {
			return fmt.Errorf("unable to insert counter: %w", err)
		}
	}

	return nil
}

func (r *dbStorage) insertGauge(ctx context.Context, db db, name string, value float64) error {
	return r.execute(
		ctx,
		db,
		`insert into metric_values (name,type,value) values ($1,$2,$3) on conflict(name,type) do update set value = $3`,
		name, models.Gauge, value,
	)
}

func (r *dbStorage) insertCounter(ctx context.Context, db db, name string, value int64) error {
	return r.execute(
		ctx,
		db,
		`insert into metric_values (name,type,delta) values ($1,$2,$3) on conflict(name,type) do update set delta = metric_values.delta + $3`,
		name, models.Counter, value,
	)
}

func (r *dbStorage) GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error) {
	metric := models.Metrics{}

	rows, err := r.query(
		ctx,
		r.pool,
		`select name, type, value, delta from metric_values where name = $1 and type = $2 limit 1`,
		mName, mType,
	)
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

	rows, err := r.query(ctx, r.pool, `select name, type, value, delta from metric_values`)
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

func (r *dbStorage) SetValues(ctx context.Context, metrics []models.Metrics) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.Sugar.Errorf("metric_values begin transaction error: %s", err.Error())
		return fmt.Errorf("metric_values begin transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, m := range metrics {

		switch m.MType {
		case models.Gauge:
			if err := r.insertGauge(ctx, tx, m.ID, *m.Value); err != nil {
				return fmt.Errorf("unable to insert gauge: %w", err)
			}
		case models.Counter:
			if err := r.insertCounter(ctx, tx, m.ID, *m.Delta); err != nil {
				return fmt.Errorf("unable to insert counter: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Sugar.Errorf("metric_values commit transaction error: %s", err.Error())
		return fmt.Errorf("metric_values commit transaction error: %w", err)
	}

	return nil
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

type db interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

func (r *dbStorage) execute(ctx context.Context, db db, query string, arguments ...any) error {
	_, err := db.Exec(ctx, query, arguments...)
	return err
}

func (r *dbStorage) query(ctx context.Context, db db, query string, arguments ...any) (pgx.Rows, error) {
	return db.Query(ctx, query, arguments...)
}
