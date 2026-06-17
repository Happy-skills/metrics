package main

import (
	"github.com/Happy-skills/metrics/internal/agent"
	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/repository"
)

func main() {
	options := config.ParseAgentFlags()
	memStore := repository.NewMemStorage()
	agent.Run(options, memStore)
}
