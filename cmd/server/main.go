package main

import (
	"context"
	"log"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/config/db"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/Happy-skills/metrics/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	options := config.ParseServerFlags()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	var store repository.Storage

	if options.DatabaseDSN != "" {
		pgxPoll, err := db.InitializeDB(ctx, options.DatabaseDSN)
		if err != nil {
			log.Println(err.Error())
		}

		store = repository.NewDBStorage(pgxPoll)
	} else {
		store = repository.NewMemStorage(options.FileStoragePath, options.StoreInterval)

		if options.RestoreOnStart && options.FileStoragePath != "" {
			repository.LoadMetricsFromFile(options, store)
		}

		if options.FileStoragePath != "" && options.StoreInterval > 0 {
			go service.WriteMetrics(options, store)
		}
	}

	if err := service.RunServer(options, store); err != nil {
		logger.Log.Fatal(err.Error())
	}
}
