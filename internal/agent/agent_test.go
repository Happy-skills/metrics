package agent

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func Test_getMetrics(t *testing.T) {
	type metric struct {
		typeMetric string
		nameMetric string
		typeValue  reflect.Kind
	}
	type want struct {
		metrics [3]metric
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test",
			want: want{metrics: [3]metric{{typeMetric: models.Counter, nameMetric: "PollCount", typeValue: reflect.Int64},
				{typeMetric: models.Gauge, nameMetric: "RandomValue", typeValue: reflect.Float64},
				{typeMetric: models.Gauge, nameMetric: "HeapAlloc", typeValue: reflect.Float64}}},
		},
	}
	mStore := repository.NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getMetrics(mStore); err != nil {
				t.Fatalf("getMetrics failed: %s", err.Error())
			}
			for i := 0; i < len(tt.want.metrics); i++ {
				v, _ := mStore.GetValue(tt.want.metrics[i].typeMetric, tt.want.metrics[i].nameMetric)
				if tt.want.metrics[i].typeMetric == models.Counter {
					if *v.Delta <= 0 {
						t.Fatalf("delta is zero for %s", tt.want.metrics[i].nameMetric)
					}
					assert.Equal(t, tt.want.metrics[i].typeValue, reflect.TypeOf(*v.Delta).Kind())
				}
				if tt.want.metrics[i].typeMetric == models.Gauge {
					if *v.Value <= 0 {
						t.Fatalf("value is zero for %s", tt.want.metrics[i].nameMetric)
					}
					assert.Equal(t, tt.want.metrics[i].typeValue, reflect.TypeOf(*v.Value).Kind())
				}
			}
		})
	}
}

func Test_sendMetrics(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "positive test",
			wantErr: false,
		},
	}
	mStore := repository.NewMemStorage()
	if err := getMetrics(mStore); err != nil {
		t.Fatalf("getMetrics failed: %s", err.Error())
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendMetrics(ts.URL, mStore); (err != nil) != tt.wantErr {
				t.Errorf("sendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
