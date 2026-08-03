package service

import (
	"net/http"
	"time"

	"github.com/Happy-skills/metrics/internal/compress"
	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/handler"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func RunServer(options config.ServerOptions, store repository.Storage) error {
	logger.Log.Info(
		"Running server",
		zap.String("addr", options.ServerAddr),
	)

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metric_type}/{metric_name}/{metric_value}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.SetMetricHandler(r.Context(), w, r, store)
		}))
		r.Post("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.SetMetricByJsonHandler(r.Context(), w, r, store)
		})))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metric_type}/{metric_name}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.GetMetricValueHandler(r.Context(), w, r, store)
		}))
		r.Post("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.GetMetricValueByJsonHandler(r.Context(), w, r, store)
		})))
	})

	r.Get("/metric/{metric_type}/{metric_name}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricHandler(r.Context(), w, r, store)
	}))

	r.Get("/ping", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
		handler.PingHandler(r.Context(), w, r, store)
	}))

	r.Get("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricsHandler(r.Context(), w, r, store)
	})))

	return http.ListenAndServe(options.ServerAddr, r)
}

func WriteMetrics(cfg config.ServerOptions, store repository.Storage) {
	ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := repository.WriteMetricsInFile(cfg.FileStoragePath, store); err != nil {
				logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
			}
		}
	}
}
