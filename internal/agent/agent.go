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
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
	metrics := make(map[string]any)
	metrics["TotalMemory"] = 0
	metrics["FreeMemory"] = 0
	metrics["CPUutilization1"] = 0

	errCh := make(chan error)
	defer close(errCh)

	dataCh := make(chan models.Metrics, options.RateLimit)
	defer close(dataCh)

	go poolGetting(ctx, options.PollInterval, memStore, dataCh, errCh)
	go poolGettingSpec(ctx, options.PollInterval, metrics, memStore, dataCh, errCh)

	for i := 1; i <= options.RateLimit; i++ {
		go workerSending("http://"+options.ServerAddr, options.Key, memStore, dataCh, errCh)
	}

	for err := range errCh {
		logger.Sugar.Errorf("Error while sending metrics: %s", err.Error())
	}

	//poolSending(ctx, options.ReportInterval, options.ServerAddr, options.Key, memStore)
}

func poolGetting(ctx context.Context, pollInterval int, memStore repository.Storage, dataCh chan models.Metrics, errCh chan error) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	fmt.Println("Starting pool getting")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Pool getting stopped")
			return
		case <-ticker.C:
			if err := getMetrics(ctx, memStore, dataCh); err != nil {
				errCh <- fmt.Errorf("error getting metrics: %s", err.Error())
			}
		}
	}
}

func poolSending(ctx context.Context, pollInterval int, flagServerAddr string, key string, memStore repository.Storage) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sendMetricsByJSONWithCompress("http://"+flagServerAddr, key, memStore); err != nil {
				logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
			}
		}
	}
}

func getMetrics(ctx context.Context, memStore repository.Storage, dataCh chan models.Metrics) error {
	if err := memStore.SetValue(ctx, models.Counter, "PollCount", int64(1)); err != nil {
		return err
	}

	met, err := memStore.GetValue(ctx, models.Counter, "PollCount")
	if err != nil {
		return err
	}
	dataCh <- *met

	if err = memStore.SetValue(ctx, models.Gauge, "RandomValue", rand.Float64()); err != nil {
		return err
	}

	met, err = memStore.GetValue(ctx, models.Gauge, "RandomValue")
	if err != nil {
		return err
	}
	dataCh <- *met

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	for k, v := range Fields(m) {
		if err := memStore.SetValue(ctx, models.Gauge, k, v); err != nil {
			return err
		}

		met, err = memStore.GetValue(ctx, models.Gauge, k)
		if err != nil {
			return err
		}
		dataCh <- *met
	}

	return nil
}

func sendMetricsByJSONWithCompress(serverURL string, key string, memStore repository.Storage) error {
	values := slices.Collect(maps.Values(memStore.GetValues(context.TODO())))
	if len(values) == 0 {
		return nil
	}

	jsonValues, err := json.Marshal(values)
	if err != nil {
		logger.Sugar.Errorf("parse json error: %s", err.Error())
		return err
	}

	err = sendDataWithRetry(
		fmt.Sprintf("%s/updates/", serverURL),
		"application/json",
		key,
		jsonValues,
	)
	if err != nil {
		logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
		return err
	}

	// reset counters after send
	for _, v := range memStore.GetValues(context.TODO()) {
		if v.ID == "PollCount" {
			if err := memStore.ResetValue(context.TODO(), models.Counter, "PollCount"); err != nil {
				logger.Sugar.Errorf("ResetValue error: %s", err.Error())
			}
		}
	}

	return nil
}

func poolGettingSpec(ctx context.Context, pollInterval int, specMetrics map[string]any, memStore repository.Storage, dataCh chan models.Metrics, errCh chan error) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := getSpecMetrics(ctx, specMetrics, memStore, dataCh); err != nil {
				errCh <- fmt.Errorf("error getting spec metrics: %s", err.Error())
			}
		}
	}
}

func getSpecMetrics(ctx context.Context, specMetrics map[string]any, memStore repository.Storage, dataCh chan models.Metrics) error {
	m, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}

	specMetrics["TotalMemory"] = m.Total
	specMetrics["FreeMemory"] = m.Free

	c, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return err
	}

	specMetrics["CPUutilization1"] = c

	for k, v := range specMetrics {
		if err = memStore.SetValue(ctx, models.Gauge, k, v); err != nil {
			return err
		}

		met, err := memStore.GetValue(ctx, models.Gauge, k)
		if err != nil {
			return err
		}
		dataCh <- *met
	}

	return nil
}

func workerSending(serverURL string, key string, memStore repository.Storage, dataCh chan models.Metrics, errCh chan error) {
	for m := range dataCh {
		err := sendMetricByJSONWithCompress(serverURL, key, m, memStore)
		if err != nil {
			errCh <- fmt.Errorf("error sending metrics: %s", err.Error())
		}
	}
}

func sendMetricByJSONWithCompress(serverURL string, key string, metric models.Metrics, memStore repository.Storage) error {
	var metrics []models.Metrics
	metrics = append(metrics, metric)

	jsonValues, err := json.Marshal(metrics)
	if err != nil {
		logger.Sugar.Errorf("parse json error: %s", err.Error())
		return err
	}

	err = sendDataWithRetry(
		fmt.Sprintf("%s/updates/", serverURL),
		"application/json",
		key,
		jsonValues,
	)
	if err != nil {
		return fmt.Errorf("sendDataWithRetry error: %s", err.Error())
	}

	// reset counters after send
	if metric.ID == "PollCount" {
		if err := memStore.ResetValue(context.TODO(), models.Counter, "PollCount"); err != nil {
			return fmt.Errorf("resetValue error: %s", err.Error())
		}
	}

	return nil
}
