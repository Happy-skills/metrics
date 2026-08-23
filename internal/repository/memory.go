package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
)

type memStorage struct {
	metrics  map[string]*models.Metrics
	filePath string
	rw       sync.RWMutex
}

func NewMemStorage(filePath string) Storage {
	return &memStorage{
		metrics:  make(map[string]*models.Metrics),
		filePath: filePath,
	}
}

func (m *memStorage) SetValue(_ context.Context, mType string, mName string, mValue any) error {
	if err := m.setValue(mType, mName, mValue); err != nil {
		return err
	}

	if m.filePath != "" {
		if err := WriteMetricsInFile(m.filePath, m); err != nil {
			logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
		}
	}

	return nil
}

func (m *memStorage) setValue(mType string, mName string, mValue any) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	var err error

	key := mType + "_" + mName
	_, ok := m.metrics[key]
	if !ok {
		m.metrics[key] = new(models.Metrics)
		m.metrics[key].MType = mType
		m.metrics[key].ID = mName
	}

	switch mType {
	case models.Gauge:
		var mVal float64
		mVal, err = toGauge(mValue)
		if err != nil {
			return err
		}

		if m.metrics[key].Value == nil {
			m.metrics[key].Value = new(float64)
		}
		*m.metrics[key].Value = mVal

		return nil
	case models.Counter:
		var mVal int64
		mVal, err = toCounter(mValue)
		if err != nil {
			return err
		}

		if m.metrics[key].Delta == nil {
			m.metrics[key].Delta = new(int64)
		}
		*m.metrics[key].Delta += mVal

		return nil
	}

	return fmt.Errorf("metrics type %q not supported", mType)
}

func (m *memStorage) SetValues(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if err := m.SetValue(ctx, metric.MType, metric.ID, *metric.Value); err != nil {
				return err
			}
		case models.Counter:
			if err := m.SetValue(ctx, metric.MType, metric.ID, *metric.Delta); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *memStorage) GetValue(_ context.Context, mType string, mName string) (*models.Metrics, error) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	key := mType + "_" + mName
	metric, ok := m.metrics[key]
	if !ok {
		return nil, errors.New("metric not found")
	}

	return metric, nil
}

func (m *memStorage) GetValues(_ context.Context) map[string]models.Metrics {
	m.rw.RLock()
	defer m.rw.RUnlock()

	ret := make(map[string]models.Metrics)
	for k, v := range m.metrics {
		ret[k] = *new(models.Metrics)
		ret[k] = *v
	}

	return ret
}

func (m *memStorage) ResetValue(ctx context.Context, mType string, mName string) error {
	return m.SetValue(ctx, mType, mName, 0)
}

func (m *memStorage) Ping(_ context.Context) error {
	return nil
}
