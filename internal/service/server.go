package service

import (
	"net/http"

	"github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

func RunServer() error {
	repository.ServerStorage = repository.NewMemStorage()

	r := chi.NewRouter()
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", handler.SetMetricHandler)
	r.Get("/get/{metric_type}/{metric_name}", handler.GetMetricHandler)
	r.Get("/value/{metric_type}/{metric_name}", handler.GetMetricValueHandler)
	r.Get("/", handler.GetMetricsHandler)

	return http.ListenAndServe(`:8080`, r)
}
