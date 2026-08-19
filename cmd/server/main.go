package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/config/db"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/Happy-skills/metrics/internal/service"
	"go.uber.org/zap"
)

func main() {
	options := config.ParseServerFlags()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	store, err := initStorage(options)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := service.RunServer(options, store); err != nil {
		logger.Log.Fatal("unable to run server", zap.Error(err))
	}
}

func initStorage(options config.ServerOptions) (repository.Storage, error) {
	if options.DatabaseDSN != "" {
		pgxPoll, err := db.InitializeDB(context.TODO(), options.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("unable to init DB: %w", err)
		}

		return repository.NewDBStorage(pgxPoll), nil
	}

	store := repository.NewMemStorage(options.FileStoragePath, options.StoreInterval)

	if options.RestoreOnStart && options.FileStoragePath != "" {
		repository.LoadMetricsFromFile(options, store)
	}

	if options.FileStoragePath != "" && options.StoreInterval > 0 {
		go service.WriteMetrics(options, store)
	}

	return store, nil
}
