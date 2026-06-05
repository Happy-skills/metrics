package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

func SetMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else {
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

		isCorrect := repository.CheckMetric(mType, mValue)
		if isCorrect == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := repository.ServerStorage.SetValue(mType, mName, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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

	m, err := repository.ServerStorage.GetValue(mType, mName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	js, jsErr := json.Marshal(m)
	if jsErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

func GetMetricValueHandler(w http.ResponseWriter, r *http.Request) {
	var mValue string
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	m, err := repository.ServerStorage.GetValue(mType, mName)
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

func GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	var strPage string

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	metrics := repository.ServerStorage.GetValues()
	if len(metrics) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	strPage = "<html><head><title>Metrics</title></head><body>\n"
	keys := make([]string, 0, len(metrics))
	for k, _ := range metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		switch metrics[k].MType {
		case models.Counter:
			strPage += "<p>" + metrics[k].ID + ": " + strconv.FormatInt(*metrics[k].Delta, 10) + "</p>\n"
		case models.Gauge:
			strPage += "<p>" + metrics[k].ID + ": " + strconv.FormatFloat(*metrics[k].Value, 'f', -1, 64) + "</p>\n"
		}
	}
	strPage += "</body></html>"

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strPage))
}
