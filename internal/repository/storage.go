package repository

import (
	"errors"
	"go/types"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
)

var ServerStorage MemStorage

type MemStorage interface {
	SetValue(mType string, mName string, mValue string) error
	GetValue(mType string, mName string) (*models.Metrics, *error)
}

type memStorage struct {
	metrics map[string]*models.Metrics
}

func NewMemStorage() MemStorage {
	return &memStorage{metrics: make(map[string]*models.Metrics)}
}

func (m *memStorage) SetValue(mType string, mName string, mValue string) error {
	key := mType + "_" + mName
	_, ok := m.metrics[key]
	if !ok {
		m.metrics[key] = new(models.Metrics)
		m.metrics[key].MType = mType
		m.metrics[key].ID = mName
	}

	if mType == models.Gauge {
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return err
		}
		if m.metrics[key].Value == nil {
			m.metrics[key].Value = new(float64)
		}
		*m.metrics[key].Value = val

		return nil
	} else if mType == models.Counter {
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return err
		}
		if m.metrics[key].Delta == nil {
			m.metrics[key].Delta = new(int64)
		}
		*m.metrics[key].Delta += val

		return nil
	}

	return types.Error{
		Msg: "metrics type not supported",
	}
}

func (m *memStorage) GetValue(mType string, mName string) (*models.Metrics, *error) {
	key := mType + "_" + mName
	metric, ok := m.metrics[key]
	if !ok {
		err := errors.New("metric not found")
		return nil, &err
	}
	return metric, nil
}
