package repository

import (
	"errors"
	"fmt"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
)

type MemStorage interface {
	SetValue(mType string, mName string, mValue any) error
	GetValue(mType string, mName string) (*models.Metrics, error)
	GetValues() map[string]models.Metrics
	ResetValue(mType string, mName string) error
}

type memStorage struct {
	metrics map[string]*models.Metrics
}

func NewMemStorage() MemStorage {
	return &memStorage{metrics: make(map[string]*models.Metrics)}
}

func (m *memStorage) SetValue(mType string, mName string, mValue any) error {
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

		return nil
	}

	return fmt.Errorf("metrics type %q not supported", mType)
}

func (m *memStorage) GetValue(mType string, mName string) (*models.Metrics, error) {
	key := mType + "_" + mName
	metric, ok := m.metrics[key]
	if !ok {
		return nil, errors.New("metric not found")
	}

	return metric, nil
}

func (m *memStorage) GetValues() map[string]models.Metrics {
	ret := make(map[string]models.Metrics)
	for k, v := range m.metrics {
		ret[k] = *new(models.Metrics)
		ret[k] = *v
	}

	return ret
}

func (m *memStorage) ResetValue(mType string, mName string) error {
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
