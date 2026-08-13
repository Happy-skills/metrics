package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Happy-skills/metrics/internal/compress"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
	"gopkg.in/h2non/gentleman.v2"
)

func sendMetrics(serverUrl string, memStore repository.Storage) error {
	var sVal string
	for _, v := range memStore.GetValues(nil) {
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

func sendMetricsByJson(serverUrl string, memStore repository.Storage) error {
	for _, v := range memStore.GetValues(nil) {
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

func sendWithRetry(url, headerValue string, body []byte) error {
	const maxRetries = 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = sendUpdateRequestWithCompress(url, headerValue, body)
		if err == nil {
			return nil
		}

		if classify(err) == NonRetriable {
			logger.Sugar.Infof("error sending metrics: %s", err.Error())
			return err
		}

		time.Sleep(time.Duration(attempt+(attempt-1)) * time.Second)
	}

	return fmt.Errorf("error sending metrics after %d attempts: %w", maxRetries, err)
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
			logger.Sugar.Errorf("Agent error gzip write: %s", err.Error())
			return err
		}
		if err := g.Close(); err != nil {
			logger.Sugar.Errorf("Agent error gzip close: %s", err.Error())
			return err
		}
		body = buf.Bytes()
		req.SetHeader("Content-Encoding", "gzip")
		req.SetHeader("Accept-Encoding", "gzip")
	}

	req.Body(bytes.NewReader(body))

	response, err := req.Send()
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	defer func(response *gentleman.Response) {
		err := response.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(response)

	if !response.Ok {
		return fmt.Errorf("agent error Send: %d %s", response.StatusCode, response.String())
	}

	return nil
}
