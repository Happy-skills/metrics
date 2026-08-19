package repository

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"slices"

	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
)

func WriteMetricsInFile(path string, memStore Storage) error {
	fd, err := os.Create(path)
	if err != nil {
		logger.Sugar.Errorf("Failed to create/open file for writing: %s", err.Error())
		return err
	}
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			logger.Sugar.Errorf("Failed to close file: %s", err.Error())
		}
	}(fd)

	metrics := slices.Collect(maps.Values(memStore.GetValues(context.TODO())))
	js, err := json.Marshal(metrics)
	if err != nil {
		logger.Sugar.Errorf("Failed to marshal metric: %s", err.Error())
		return err
	}
	_, err = fd.Write(js)
	if err != nil {
		logger.Sugar.Errorf("Failed to write metric in the file: %s", err.Error())
		return err
	}

	return nil
}

func LoadMetricsFromFile(path string, memStore Storage) {
	var metrics []models.Metrics

	fd, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Sugar.Errorf("Failed to open file : %s", err.Error())
	}
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			logger.Sugar.Errorf("Failed to close file: %s", err.Error())
		}
	}(fd)

	if err := json.NewDecoder(fd).Decode(&metrics); err != nil {
		logger.Sugar.Errorf("Failed to unmarshal metrics from file: %s", err.Error())
		return
	}

	if err := memStore.SetValues(context.TODO(), metrics); err != nil {
		logger.Sugar.Errorf("Failed to load metric to store: %s", err.Error())
	}
}
