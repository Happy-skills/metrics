package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Happy-skills/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestSetMetricHandler(t *testing.T) {
	type want struct {
		code             int
		requestMethod    string
		requestString    string
		parametersString map[string]string
		msgString        string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test",
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "POST update/counter/PollCount/10",
			},
		},
		{
			name: "wrong method Get",
			want: want{
				code:             http.StatusMethodNotAllowed,
				requestMethod:    http.MethodGet,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "GET update/counter/PollCount/10",
			},
		},
		{
			name: "wrong method Put",
			want: want{
				code:             http.StatusMethodNotAllowed,
				requestMethod:    http.MethodPut,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "PUT update/counter/PollCount/10",
			},
		},
		{
			name: "empty value",
			want: want{
				code:             http.StatusBadRequest,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount"},
				msgString:        "POST update/counter/PollCount",
			},
		},
		{
			name: "wrong value",
			want: want{
				code:             http.StatusBadRequest,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount", "metric_value": "10.953644889"},
				msgString:        "POST update/counter/PollCount/10.953644889",
			},
		},
		{
			name: "empty name",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_value": "10"},
				msgString:        "POST update/counter/10",
			},
		},
		{
			name: "wrong type",
			want: want{
				code:             http.StatusBadRequest,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "count", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "POST update/count/PollCount/10",
			},
		},
	}
	repository.ServerStorage = repository.NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString+"/{metric_type}/{metric_name}/{metric_value}", nil)
			for k, v := range tt.want.parametersString {
				request.SetPathValue(k, v)
			}
			w := httptest.NewRecorder()
			SetMetricHandler(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
		})
	}
}
