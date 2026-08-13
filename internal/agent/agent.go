package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"runtime"
	"slices"
	"time"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

func Fields(r runtime.MemStats) map[string]any {
	return map[string]any{
		"Alloc":         r.Alloc,
		"TotalAlloc":    r.TotalAlloc,
		"Sys":           r.Sys,
		"Lookups":       r.Lookups,
		"Mallocs":       r.Mallocs,
		"Frees":         r.Frees,
		"HeapAlloc":     r.HeapAlloc,
		"HeapSys":       r.HeapSys,
		"HeapIdle":      r.HeapIdle,
		"HeapInuse":     r.HeapInuse,
		"HeapReleased":  r.HeapReleased,
		"HeapObjects":   r.HeapObjects,
		"StackInuse":    r.StackInuse,
		"StackSys":      r.StackSys,
		"MSpanInuse":    r.MSpanInuse,
		"MSpanSys":      r.MSpanSys,
		"MCacheInuse":   r.MCacheInuse,
		"MCacheSys":     r.MCacheSys,
		"BuckHashSys":   r.BuckHashSys,
		"GCSys":         r.GCSys,
		"OtherSys":      r.OtherSys,
		"NextGC":        r.NextGC,
		"LastGC":        r.LastGC,
		"PauseTotalNs":  r.PauseTotalNs,
		"NumGC":         r.NumGC,
		"NumForcedGC":   r.NumForcedGC,
		"GCCPUFraction": r.GCCPUFraction,
	}
}

func Run(ctx context.Context, options config.AgentOptions, memStore repository.Storage) {
	go poolGetting(ctx, options.PollInterval, memStore)
	poolSending(ctx, options.ReportInterval, options.ServerAddr, memStore)
}

func poolGetting(ctx context.Context, pollInterval int, memStore repository.Storage) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := getMetrics(ctx, memStore); err != nil {
				logger.Sugar.Errorf("Error getting metrics: %s", err.Error())
			}
		}
	}
}

func poolSending(ctx context.Context, pollInterval int, flagServerAddr string, memStore repository.Storage) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sendMetricsByJsonWithCompress("http://"+flagServerAddr, memStore); err != nil {
				logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
			}
		}
	}
}

func getMetrics(ctx context.Context, memStore repository.Storage) error {
	if err := memStore.SetValue(ctx, models.Counter, "PollCount", int64(1)); err != nil {
		return err
	}
	if err := memStore.SetValue(ctx, models.Gauge, "RandomValue", rand.Float64()); err != nil {
		return err
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	for k, v := range Fields(m) {
		if err := memStore.SetValue(ctx, models.Gauge, k, v); err != nil {
			return err
		}
	}

	return nil
}

func sendMetricsByJsonWithCompress(serverUrl string, memStore repository.Storage) error {
	values := slices.Collect(maps.Values(memStore.GetValues(nil)))
	if len(values) == 0 {
		return nil
	}

	jsonValues, err := json.Marshal(values)
	if err != nil {
		logger.Sugar.Errorf("parse json error: %s", err.Error())
		return err
	}

	err = sendWithRetry(
		fmt.Sprintf("%s/updates/", serverUrl),
		"application/json",
		jsonValues,
	)
	if err != nil {
		logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
		return err
	}

	// reset counters after send
	for _, v := range memStore.GetValues(nil) {
		if v.ID == "PollCount" {
			if err := memStore.ResetValue(models.Counter, "PollCount"); err != nil {
				logger.Sugar.Errorf("ResetValue error: %s", err.Error())
			}
		}
	}

	return nil
}
