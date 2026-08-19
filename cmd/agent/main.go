package main

import (
	"log"

	"github.com/Happy-skills/metrics/internal/agent"
	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/Happy-skills/metrics/internal/retrier"
	"golang.org/x/net/context"
)

func main() {
	options := config.ParseAgentFlags()
	memStore := repository.NewMemStorage("", 0)

	ctx := context.Background()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err.Error())
	}

	retrier.SetRetries(options.SendRetries)

	agent.Run(ctx, options, memStore)
}
