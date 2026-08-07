package handler

import (
	"context"
	"fmt"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

func updateMetric(ctx context.Context, metric *models.Metrics, store repository.Storage) error {
	var mValue string
	if metric.Delta != nil {
		mValue = strconv.FormatInt(*metric.Delta, 10)
	} else if metric.Value != nil {
		mValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	if err := repository.CheckMetric(metric.MType, mValue); err != nil {
		return fmt.Errorf("%w: check metrics failed", err)
	}

	if err := store.SetValue(ctx, metric.MType, metric.ID, mValue); err != nil {
		return fmt.Errorf("%w: set metrics failed", err)
	}

	return nil
}
