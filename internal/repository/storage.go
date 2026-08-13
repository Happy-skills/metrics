package repository

import (
	"context"

	models "github.com/Happy-skills/metrics/internal/model"
)

type Storage interface {
	SetValue(ctx context.Context, mType string, mName string, mValue any) error
	GetValue(ctx context.Context, mType string, mName string) (*models.Metrics, error)
	GetValues(ctx context.Context) map[string]models.Metrics
	SetValues(ctx context.Context, models []models.Metrics) error
	ResetValue(mType string, mName string) error
	GetValuesSlice() []models.Metrics
	SetValuesFromSlice(metrics []models.Metrics) error
	Ping(ctx context.Context) error
}
