package handler

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

func SetMetricHandler(w http.ResponseWriter, r *http.Request, cfg config.ServerOptions, memStore repository.MemStorage) {
	mName := r.PathValue("metric_name")
	if mName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mType := r.PathValue("metric_type")
	if mType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mValue := r.PathValue("metric_value")

	if err := repository.CheckMetric(mType, mValue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := memStore.SetValue(mType, mName, mValue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if cfg.StoreInterval == 0 {
		if err := repository.WriteMetricsInFile(cfg, memStore); err != nil {
			logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request, memStore repository.MemStorage) {
	mType := r.PathValue("metric_type")
	if mType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mName := r.PathValue("metric_name")
	if mName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	m, err := memStore.GetValue(mType, mName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	js, jsErr := json.Marshal(m)
	if jsErr != nil {
		logger.Sugar.Errorf("Failed to marshal metric value %s", jsErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

func GetMetricValueHandler(w http.ResponseWriter, r *http.Request, memStore repository.MemStorage) {
	var mValue string
	mType := r.PathValue("metric_type")
	if mType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mName := r.PathValue("metric_name")
	if mName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	m, err := memStore.GetValue(mType, mName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch mType {
	case models.Counter:
		mValue = strconv.FormatInt(*m.Delta, 10)
	case models.Gauge:
		mValue = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(mValue))
}

func GetMetricsHandler(w http.ResponseWriter, r *http.Request, memStore repository.MemStorage) {
	var dataHTML []string
	const tpl = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Metrics</title>
	</head>
	<body>
		{{range $m := .}}<div>{{$m}}</div>{{else}}<div></div>{{end}}
	</body>
</html>`

	tmpl, err := template.New("metricsHTML").Parse(tpl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	metrics := memStore.GetValues()

	keys := make([]string, 0, len(metrics))
	for k := range metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		switch metrics[k].MType {
		case models.Counter:
			dataHTML = append(dataHTML, metrics[k].ID+": "+strconv.FormatInt(*metrics[k].Delta, 10))
		case models.Gauge:
			dataHTML = append(dataHTML, metrics[k].ID+": "+strconv.FormatFloat(*metrics[k].Value, 'f', -1, 64))
		}
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "metricsHTML", dataHTML)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

func SetMetricByJsonHandler(w http.ResponseWriter, r *http.Request, cfg config.ServerOptions, memStore repository.MemStorage) {
	var metric models.Metrics

	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar.Errorf("Error reading body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(jsonBody, &metric); err != nil {
		logger.Sugar.Errorf("Error parsing body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if metric.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var mValue string
	if metric.Delta != nil {
		mValue = strconv.FormatInt(*metric.Delta, 10)
	} else if metric.Value != nil {
		mValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	if err := repository.CheckMetric(metric.MType, mValue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := memStore.SetValue(metric.MType, metric.ID, mValue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if cfg.StoreInterval == 0 {
		if err := repository.WriteMetricsInFile(cfg, memStore); err != nil {
			logger.Sugar.Errorf("Error writting metrics in file: %s", err.Error())
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func GetMetricValueByJsonHandler(w http.ResponseWriter, r *http.Request, memStore repository.MemStorage) {
	var reqMetric models.Metrics

	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar.Errorf("Error reading body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(jsonBody, &reqMetric); err != nil {
		logger.Sugar.Errorf("Error parsing body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if reqMetric.MType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if reqMetric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metric, err := memStore.GetValue(reqMetric.MType, reqMetric.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	js, jsErr := json.Marshal(metric)
	if jsErr != nil {
		logger.Sugar.Errorf("Failed to marshal metric value %s", jsErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}
