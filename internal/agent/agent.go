package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Happy-skills/metrics/internal/compress"
	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	"gopkg.in/h2non/gentleman.v2"

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

func Run(options config.AgentOptions, memStore repository.MemStorage) {
	go poolGetting(options.PollInterval, memStore)
	poolSending(options.ReportInterval, options.ServerAddr, memStore)
}

func poolGetting(pollInterval int, memStore repository.MemStorage) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := getMetrics(memStore); err != nil {
				logger.Sugar.Errorf("Error getting metrics: %s", err.Error())
			}
		}
	}
}

func poolSending(pollInterval int, flagServerAddr string, memStore repository.MemStorage) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendMetricsByJsonWithCompress("http://"+flagServerAddr, memStore); err != nil {
				logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
			}
		}
	}
}

func getMetrics(memStore repository.MemStorage) error {
	if err := memStore.SetValue(models.Counter, "PollCount", int64(1)); err != nil {
		return err
	}
	if err := memStore.SetValue(models.Gauge, "RandomValue", rand.Float64()); err != nil {
		return err
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	for k, v := range Fields(m) {
		if err := memStore.SetValue(models.Gauge, k, v); err != nil {
			return err
		}
	}

	return nil
}

func sendMetrics(serverUrl string, memStore repository.MemStorage) error {
	var sVal string
	for _, v := range memStore.GetValues() {
		vType := v.MType
		switch vType {
		case models.Counter:
			sVal = strconv.FormatInt(*v.Delta, 10)
		case models.Gauge:
			sVal = strconv.FormatFloat(*v.Value, 'f', -1, 64)
		}
		//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		url := fmt.Sprintf("%s/update/%s/%s/%s", serverUrl, v.MType, v.ID, sVal)
		if err := sendUpdateRequest(url, "text/plain", ""); err != nil {
			logger.Sugar.Fatalf("sendMetrics error: %s", err.Error())
		}

		if v.ID == "PollCount" {
			if err := memStore.ResetValue(models.Counter, "PollCount"); err != nil {
				logger.Sugar.Fatalf("ResetValue error: %s", err.Error())
			}
		}
	}

	return nil
}

func sendUpdateRequest(url string, headerValue string, body string) error {
	client := gentleman.New()
	req := client.Request()
	req.Method(http.MethodPost)
	req.URL(url)
	req.SetHeader("Content-Type", headerValue)
	req.Body(strings.NewReader(body))

	response, err := req.Send()
	if err != nil {
		return err
	}
	defer func(response *gentleman.Response) {
		err := response.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(response)

	if !response.Ok {
		return errors.New(response.String())
	}

	return nil
}

func sendMetricsByJson(serverUrl string, memStore repository.MemStorage) error {
	for _, v := range memStore.GetValues() {
		url := fmt.Sprintf("%s/update", serverUrl)
		jsonValue, err := json.Marshal(v)
		if err != nil {
			logger.Sugar.Errorf("parse json error: %s", err.Error())
		}
		if err := sendUpdateRequest(url, "application/json", string(jsonValue)); err != nil {
			logger.Sugar.Errorf("sendMetrics error: %s", err.Error())
		}

		if v.ID == "PollCount" {
			if err := memStore.ResetValue(models.Counter, "PollCount"); err != nil {
				logger.Sugar.Errorf("ResetValue error: %s", err.Error())
			}
		}
	}

	return nil
}

func sendUpdateRequestWithCompress(url string, headerValue string, body []byte) error {
	client := gentleman.New()
	req := client.Request()
	req.Method(http.MethodPost)
	req.URL(url)
	req.SetHeader("Content-Type", headerValue)

	contentType := slices.Contains(compress.TypesForGzip, headerValue)
	if contentType {
		var buf bytes.Buffer
		defer buf.Reset()

		g := gzip.NewWriter(&buf)
		if _, err := g.Write(body); err != nil {
			logger.Sugar.Errorf("Agent error gzip write: %T %+v", err, err)
			return err
		}
		if err := g.Close(); err != nil {
			logger.Sugar.Errorf("Agent error gzip close: %T %+v", err, err)
			return err
		}
		body = buf.Bytes()
		req.SetHeader("Content-Encoding", "gzip")
		req.SetHeader("Accept-Encoding", "gzip")
	}

	req.Body(bytes.NewReader(body))

	response, err := req.Send()
	if err != nil {
		logger.Sugar.Errorf("Agent error Send: %T %+v", err, err)
		return err
	}
	defer func(response *gentleman.Response) {
		err := response.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(response)

	if !response.Ok {
		return errors.New(strconv.Itoa(response.StatusCode))
	}

	return nil
}

func sendMetricsByJsonWithCompress(serverUrl string, memStore repository.MemStorage) error {
	for _, v := range memStore.GetValues() {
		url := fmt.Sprintf("%s/update", serverUrl)
		jsonValue, err := json.Marshal(v)
		if err != nil {
			logger.Sugar.Errorf("parse json error: %s", err.Error())
			return err
		}
		if err := sendUpdateRequestWithCompress(url, "application/json", jsonValue); err != nil {
			logger.Sugar.Errorf("sendUpdateRequest error: %s", err.Error())
			return err
		}

		if v.ID == "PollCount" {
			if err := memStore.ResetValue(models.Counter, "PollCount"); err != nil {
				logger.Sugar.Errorf("ResetValue error: %s", err.Error())
			}
		}
	}

	return nil
}
