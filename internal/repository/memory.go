package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
)

type memStorage struct {
	metrics      map[string]*models.Metrics
	filePath     string
	fileInterval int
	mu           sync.Mutex
}

func NewMemStorage(filePath string, fileInterval int) Storage {
	return &memStorage{metrics: make(map[string]*models.Metrics), filePath: filePath, fileInterval: fileInterval}
}

func (m *memStorage) SetValue(ctx context.Context, mType string, mName string, mValue any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	key := mType + "_" + mName
	_, ok := m.metrics[key]
	if !ok {
		m.metrics[key] = new(models.Metrics)
		m.metrics[key].MType = mType
		m.metrics[key].ID = mName
	}

	if mType == models.Gauge {
		var mVal float64

		switch mValue.(type) {
		case uint64:
			mVal = float64(mValue.(uint64))
		case uint32:
			mVal = float64(mValue.(uint32))
		case int64:
			mVal = float64(mValue.(int64))
		case float64:
			mVal = mValue.(float64)
		case string:
			mVal, err = strconv.ParseFloat(mValue.(string), 64)
			if err != nil {
				return err
			}
		}

		if m.metrics[key].Value == nil {
			m.metrics[key].Value = new(float64)
		}
		*m.metrics[key].Value = mVal

		if m.filePath != "" {
			if err := WriteMetricsInFile(m.filePath, m); err != nil {
				logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
			}
		}

		return nil

	} else if mType == models.Counter {
		var mVal int64

		switch mValue.(type) {
		case uint64:
			mVal = int64(mValue.(uint64))
		case uint32:
			mVal = int64(mValue.(uint32))
		case float64:
			mVal = int64(mValue.(float64))
		case int64:
			mVal = mValue.(int64)
		case string:
			mVal, err = strconv.ParseInt(mValue.(string), 10, 64)
			if err != nil {
				return err
			}
		}

		if m.metrics[key].Delta == nil {
			m.metrics[key].Delta = new(int64)
		}
		*m.metrics[key].Delta += mVal

		if m.filePath != "" {
			if err := WriteMetricsInFile(m.filePath, m); err != nil {
				logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
			}
		}

		return nil
	}

	return fmt.Errorf("metrics type %q not supported", mType)
}

func (m *memStorage) GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error) {
	key := mType + "_" + mName
	metric, ok := m.metrics[key]
	if !ok {
		return nil, errors.New("metric not found")
	}

	return metric, nil
}

func (m *memStorage) GetValues(ctx context.Context) map[string]models.Metrics {
	ret := make(map[string]models.Metrics)
	for k, v := range m.metrics {
		ret[k] = *new(models.Metrics)
		ret[k] = *v
	}

	return ret
}

func (m *memStorage) ResetValue(mType string, mName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := mType + "_" + mName
	_, ok := m.metrics[key]
	if !ok {
		m.metrics[key] = new(models.Metrics)
		m.metrics[key].MType = mType
		m.metrics[key].ID = mName
	}

	if mType == models.Gauge {
		if m.metrics[key].Value == nil {
			m.metrics[key].Value = new(float64)
		}
		*m.metrics[key].Value = 0

		return nil

	} else if mType == models.Counter {
		if m.metrics[key].Delta == nil {
			m.metrics[key].Delta = new(int64)
		}
		*m.metrics[key].Delta = 0

		return nil
	}

	return fmt.Errorf("metrics type %q not supported", mType)
}

func (m *memStorage) GetValuesSlice() []models.Metrics {
	var ret []models.Metrics
	for _, v := range m.metrics {
		ret = append(ret, *v)
	}
	return ret
}

func (m *memStorage) SetValuesFromSlice(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if metric.Value != nil {
			if err := m.SetValue(nil, metric.MType, metric.ID, *metric.Value); err != nil {
				return err
			}
		}
		if metric.Delta != nil {
			if err := m.SetValue(nil, metric.MType, metric.ID, *metric.Delta); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *memStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *memStorage) Begin(ctx context.Context) error {
	m.mu.Lock()
	return nil
}

func (m *memStorage) Commit(ctx context.Context) error {
	m.mu.Unlock()
	return nil
}

func (m *memStorage) Rollback(ctx context.Context) error {
	m.mu.Unlock()
	return nil
}
