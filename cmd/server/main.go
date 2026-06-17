package main

import (
	"log"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/Happy-skills/metrics/internal/service"
)

func main() {
	options := config.ParseServerFlags()
	memStore := repository.NewMemStorage()

	if err := service.RunServer(options, memStore); err != nil {
		log.Fatal(err)
	}
}
