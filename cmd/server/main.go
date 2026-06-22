package main

import (
	"log"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/Happy-skills/metrics/internal/service"
)

func main() {
	options := config.ParseServerFlags()
	memStore := repository.NewMemStorage()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	if err := service.RunServer(options, memStore); err != nil {
		logger.Log.Fatal(err.Error())
	}
}
