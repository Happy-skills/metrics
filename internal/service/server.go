package service

import (
	"log"
	"net/http"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

func RunServer(options config.ServerOptions, memStore repository.MemStorage) error {
	log.Println("Running server ", options.ServerAddr)

	r := chi.NewRouter()
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", func(w http.ResponseWriter, r *http.Request) {
		handler.SetMetricHandler(w, r, memStore)
	})
	r.Get("/get/{metric_type}/{metric_name}", func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricHandler(w, r, memStore)
	})
	r.Get("/value/{metric_type}/{metric_name}", func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricValueHandler(w, r, memStore)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricsHandler(w, r, memStore)
	})

	return http.ListenAndServe(options.ServerAddr, r)
}
