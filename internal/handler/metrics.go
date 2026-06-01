package handler

import (
	"encoding/json"
	"net/http"

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
