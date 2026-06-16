package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
)

func SetMetricHandler(w http.ResponseWriter, r *http.Request, memStore repository.MemStorage) {
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
	const tpl = `
<!DOCTYPE html>
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

	err = tmpl.ExecuteTemplate(w, "metricsHTML", dataHTML)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}
