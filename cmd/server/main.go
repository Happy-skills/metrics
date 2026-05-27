package main

import (
	_ "github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/service"
)

func main() {
	if err := service.RunServer(); err != nil {
		panic(err)
	}
}
