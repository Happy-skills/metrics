package main

import (
	"github.com/Happy-skills/metrics/internal/config"
	_ "github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/service"
)

func main() {
	config.ParseServerFlags()

	if err := service.RunServer(); err != nil {
		panic(err)
	}
}
