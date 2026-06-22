package main

import (
	"log"

	"github.com/Happy-skills/metrics/internal/agent"
	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
)

func main() {
	options := config.ParseAgentFlags()
	memStore := repository.NewMemStorage()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	agent.Run(options, memStore)
}
