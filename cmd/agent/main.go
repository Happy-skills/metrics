package main

import (
	"github.com/Happy-skills/metrics/internal/agent"
	"github.com/Happy-skills/metrics/internal/config"
)

func main() {
	config.ParseAgentFlags()
	agent.Run()
}
