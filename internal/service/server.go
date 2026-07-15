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

func RunServer(options config.ServerOptions, memStore repository.MemStorage) error {
	logger.Log.Info(
		"Running server",
		zap.String("addr", options.ServerAddr),
	)

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metric_type}/{metric_name}/{metric_value}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.SetMetricHandler(w, r, options, memStore)
		}))
		r.Post("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.SetMetricByJsonHandler(w, r, options, memStore)
		})))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metric_type}/{metric_name}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.GetMetricValueHandler(w, r, memStore)
		}))
		r.Post("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
			handler.GetMetricValueByJsonHandler(w, r, memStore)
		})))
	})

	r.Get("/get/{metric_type}/{metric_name}", logger.LoggingHandler(func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricHandler(w, r, memStore)
	}))

	r.Get("/", logger.LoggingHandler(compress.GzipHandler(func(w http.ResponseWriter, r *http.Request) {
		handler.GetMetricsHandler(w, r, memStore)
	})))

	return http.ListenAndServe(options.ServerAddr, r)
}

func WriteMetrics(cfg config.ServerOptions, memStore repository.MemStorage) {
	ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := repository.WriteMetricsInFile(cfg, memStore); err != nil {
				logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
			}
		}
	}
}
