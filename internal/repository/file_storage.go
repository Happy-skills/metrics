package repository

import (
	"encoding/json"
	"os"

	"github.com/Happy-skills/metrics/internal/config"
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

	metrics := memStore.GetValuesSlice()
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

func LoadMetricsFromFile(cfg config.ServerOptions, memStore Storage) {
	var metrics []models.Metrics

	fd, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
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

	if err := memStore.SetValuesFromSlice(metrics); err != nil {
		logger.Sugar.Errorf("Failed to store metric: %s", err.Error())
	}
}
