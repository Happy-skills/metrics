package service

import (
	"net/http"

	"github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/repository"
)

func RunServer() error {
	repository.ServerStorage = repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handler.SetMetricHandler)
	mux.HandleFunc("/get/{metric_type}/{metric_name}", handler.GetMetricHandler)
	mux.HandleFunc("/", handler.MainHandler)

	return http.ListenAndServe(`:8080`, mux)
}
