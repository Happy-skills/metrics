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
	memStore := repository.NewMemStorage()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	pgxPoll, err := db.InitializeDB(ctx, options.DatabaseDSN)
	if err != nil {
		log.Fatal(err.Error())
	}

	repo := repository.New(pgxPoll)

	if options.RestoreOnStart {
		repository.LoadMetricsFromFile(options, memStore)
	}

	if options.StoreInterval > 0 {
		go service.WriteMetrics(options, memStore)
	}

	if err := service.RunServer(options, memStore, repo); err != nil {
		logger.Log.Fatal(err.Error())
	}
}
