package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Happy-skills/metrics/internal/config"
	"github.com/Happy-skills/metrics/internal/logger"
	"github.com/Happy-skills/metrics/internal/mocks"
	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		cfg  config.ServerOptions
		want want
	}{
		{
			name: "positive test",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "POST update/counter/PollCount/10",
			},
		},
		{
			name: "empty value",
			cfg:  config.ServerOptions{StoreInterval: 60},
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
			cfg:  config.ServerOptions{StoreInterval: 60},
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
			cfg:  config.ServerOptions{StoreInterval: 60},
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
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:             http.StatusBadRequest,
				requestMethod:    http.MethodPost,
				requestString:    "update",
				parametersString: map[string]string{"metric_type": "count", "metric_name": "PollCount", "metric_value": "10"},
				msgString:        "POST update/count/PollCount/10",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString+"/{metric_type}/{metric_name}/{metric_value}", nil)
			for k, v := range tt.want.parametersString {
				request.SetPathValue(k, v)
			}
			w := httptest.NewRecorder()

			if tt.want.code == http.StatusOK {
				mockStorage.EXPECT().SetValue(
					gomock.Any(),
					tt.want.parametersString["metric_type"],
					tt.want.parametersString["metric_name"],
					tt.want.parametersString["metric_value"],
				).Return(nil).Times(1)
			}
			SetMetricHandler(t.Context(), w, request, mockStorage)

			res := w.Result()
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	type want struct {
		code             int
		requestMethod    string
		requestString    string
		parametersString map[string]string
		containsValue    string
		contentType      string
		msgString        string
		mock             func(m *mocks.MockStorage)
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive counter",
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount"},
				containsValue:    "PollCount",
				contentType:      "application/json",
				msgString:        "GET get/counter/PollCount",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "counter", "PollCount").
						Return(
							&models.Metrics{
								ID:    "PollCount",
								MType: "counter",
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "positive gauge",
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "gauge", "metric_name": "RandomValue"},
				containsValue:    "RandomValue",
				contentType:      "application/json",
				msgString:        "GET get/gauge/RandomValue",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "RandomValue").
						Return(
							&models.Metrics{
								ID:    "RandomValue",
								MType: "gauge",
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "empty type",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_name": "PollCount"},
				containsValue:    "",
				msgString:        "GET get/PollCount",
				mock:             func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong type misprint",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "count", "metric_name": "PollCount"},
				containsValue:    "",
				msgString:        "GET get/count/PollCount",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "count", "PollCount").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},
		{
			name: "wrong type gauge for PollCount",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "gauge", "metric_name": "PollCount"},
				containsValue:    "",
				msgString:        "GET get/gauge/PollCount",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "PollCount").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},
		{
			name: "empty name",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "counter"},
				containsValue:    "",
				msgString:        "GET get/counter",
				mock:             func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong name",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "get",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCounter"},
				containsValue:    "",
				msgString:        "GET get/count/PollCounter",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "counter", "PollCounter").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	if err := logger.Initialize("info"); err != nil {
		t.Fatal(err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString+"/{metric_type}/{metric_name}", nil)
			for k, v := range tt.want.parametersString {
				request.SetPathValue(k, v)
			}
			tt.want.mock(mockStorage)

			w := httptest.NewRecorder()

			GetMetricHandler(t.Context(), w, request, mockStorage)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			assert.Contains(t, string(resBody), tt.want.containsValue)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetricValueHandler(t *testing.T) {
	type want struct {
		code             int
		requestMethod    string
		requestString    string
		parametersString map[string]string
		contentType      string
		msgString        string
		mock             func(m *mocks.MockStorage)
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive counter",
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "counter", "metric_name": "PollCount"},
				contentType:      "text/plain",
				msgString:        "GET value/counter/PollCount",
				mock: func(m *mocks.MockStorage) {
					value := int64(2)

					m.EXPECT().
						GetValue(gomock.Any(), "counter", "PollCount").
						Return(
							&models.Metrics{
								ID:    "PollCount",
								MType: "counter",
								Delta: &value,
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "positive gauge",
			want: want{
				code:             http.StatusOK,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "gauge", "metric_name": "RandomValue"},
				contentType:      "text/plain",
				msgString:        "GET value/gauge/RandomValue",
				mock: func(m *mocks.MockStorage) {
					value := float64(2)

					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "RandomValue").
						Return(
							&models.Metrics{
								ID:    "RandomValue",
								MType: "gauge",
								Value: &value,
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "empty type",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_name": "PollCount"},
				msgString:        "GET value/PollCount",
				mock:             func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong type misprint",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "count", "metric_name": "PollCount"},
				msgString:        "GET value/count/PollCount",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "count", "PollCount").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},

		{
			name: "wrong type gauge for PollCount",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "gauge", "metric_name": "PollCount"},
				msgString:        "GET value/gauge/PollCount",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "PollCount").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},
		{
			name: "empty name",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "counter"},
				msgString:        "GET value/counter",
				mock:             func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong name",
			want: want{
				code:             http.StatusNotFound,
				requestMethod:    http.MethodGet,
				requestString:    "value",
				parametersString: map[string]string{"metric_type": "count", "metric_name": "PollCounter"},
				msgString:        "GET value/count/PollCounter",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "count", "PollCounter").
						Return(
							nil,
							errors.New("metric not found"),
						).
						Times(1)
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString+"/{metric_type}/{metric_name}", nil)
			for k, v := range tt.want.parametersString {
				request.SetPathValue(k, v)
			}

			tt.want.mock(mockStorage)

			w := httptest.NewRecorder()

			GetMetricValueHandler(t.Context(), w, request, mockStorage)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			if res.StatusCode == 200 && !assert.NotEmpty(t, resBody) {
				t.Fatalf("empty body with response status code 200 for %s", tt.want.msgString)
			}
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetricsHandler(t *testing.T) {
	type want struct {
		code          int
		requestMethod string
		contentType   string
		bodyString    []string
		msgString     string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test",
			want: want{
				code:          http.StatusOK,
				requestMethod: http.MethodGet,
				contentType:   "text/html",
				bodyString:    []string{"PollCount", "RandomValue"},
				msgString:     "GET /",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/", nil)

			delta := int64(2)
			value := float64(2)

			mockStorage.EXPECT().
				GetValues(gomock.Any()).
				Return(
					map[string]models.Metrics{
						"PollCount": {
							ID:    "PollCount",
							MType: "counter",
							Delta: &delta,
						},
						"RandomValue": {
							ID:    "RandomValue",
							MType: "gauge",
							Value: &value,
						},
					},
				).
				Times(1)

			w := httptest.NewRecorder()

			GetMetricsHandler(t.Context(), w, request, mockStorage)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
			resBody, err := io.ReadAll(res.Body)
			t.Log(string(resBody))
			require.NoError(t, err)
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			if res.StatusCode == 200 {
				if !assert.NotEmpty(t, resBody) {
					t.Fatalf("empty body with response status code 200 for %s", tt.want.msgString)
				}
				for i := 0; i < len(tt.want.bodyString); i++ {
					assert.Contains(t, string(resBody), tt.want.bodyString[i])
				}
			}
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestSetMetricByJsonHandler(t *testing.T) {
	type want struct {
		code           int
		requestMethod  string
		requestString  string
		parametersJson []byte
		msgString      string
		mock           func(m *mocks.MockStorage)
	}
	tests := []struct {
		name string
		cfg  config.ServerOptions
		want want
	}{
		{
			name: "positive counter",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusOK,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"id":"PollCount", "type":"counter", "delta":10}`),
				msgString:      "POST update",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().SetValue(gomock.Any(), "counter", "PollCount", "10").
						Return(nil).
						Times(1)
				},
			},
		},
		{
			name: "positive gauge",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusOK,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"id":"RandomValue", "type":"gauge", "value":1.65892}`),
				msgString:      "POST update",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().SetValue(gomock.Any(), "gauge", "RandomValue", "1.65892").
						Return(nil).
						Times(1)
				},
			},
		},
		{
			name: "empty value",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusBadRequest,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"id":"PollCount", "type":"counter"}`),
				msgString:      "POST update",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong value",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusInternalServerError,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"id":"PollCount", "type":"counter", "delta":10.953644889}`),
				msgString:      "POST update",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "empty name",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusBadRequest,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"type":"counter", "value":10}`),
				msgString:      "POST update",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong type",
			cfg:  config.ServerOptions{StoreInterval: 60},
			want: want{
				code:           http.StatusBadRequest,
				requestMethod:  http.MethodPost,
				requestString:  "update",
				parametersJson: []byte(`{"id":"PollCount", "type":"count", "delta":10}`),
				msgString:      "POST update",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	if err := logger.Initialize("info"); err != nil {
		t.Fatal(err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString, bytes.NewBuffer(tt.want.parametersJson))

			tt.want.mock(mockStorage)

			w := httptest.NewRecorder()

			SetMetricByJsonHandler(t.Context(), w, request, mockStorage)
			res := w.Result()
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
		})
	}
}

func TestGetMetricValueByJsonHandler(t *testing.T) {
	type want struct {
		code             int
		requestMethod    string
		requestString    string
		parametersString map[string]string
		parametersJson   []byte
		containsValue    string
		contentType      string
		msgString        string
		mock             func(m *mocks.MockStorage)
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive counter",
			want: want{
				code:           http.StatusOK,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"PollCount", "type":"counter"}`),
				containsValue:  "PollCount",
				contentType:    "application/json",
				msgString:      "POST value",
				mock: func(m *mocks.MockStorage) {
					value := int64(2)

					m.EXPECT().
						GetValue(gomock.Any(), "counter", "PollCount").
						Return(
							&models.Metrics{
								ID:    "PollCount",
								MType: "counter",
								Delta: &value,
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "positive gauge",
			want: want{
				code:           http.StatusOK,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"RandomValue", "type":"gauge"}`),
				containsValue:  "RandomValue",
				contentType:    "application/json",
				msgString:      "POST value",
				mock: func(m *mocks.MockStorage) {
					value := float64(2)

					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "RandomValue").
						Return(
							&models.Metrics{
								ID:    "RandomValue",
								MType: "gauge",
								Value: &value,
							},
							nil,
						).
						Times(1)
				},
			},
		},
		{
			name: "empty type",
			want: want{
				code:           http.StatusNotFound,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"PollCount"}`),
				containsValue:  "",
				msgString:      "POST value",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong type misprint",
			want: want{
				code:           http.StatusNotFound,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"PollCount", "type":"count"}`),
				containsValue:  "",
				msgString:      "POST value",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "count", "PollCount").
						Return(
							nil,
							errors.New("not found"),
						).
						Times(1)
				},
			},
		},
		{
			name: "wrong type gauge for PollCount",
			want: want{
				code:           http.StatusNotFound,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"PollCount", "type":"gauge"}`),
				containsValue:  "",
				msgString:      "POST value",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "gauge", "PollCount").
						Return(
							nil,
							errors.New("not found"),
						).
						Times(1)
				},
			},
		},
		{
			name: "empty name",
			want: want{
				code:           http.StatusNotFound,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"type":"counter"}`),
				containsValue:  "",
				msgString:      "POST value",
				mock:           func(m *mocks.MockStorage) {},
			},
		},
		{
			name: "wrong name",
			want: want{
				code:           http.StatusNotFound,
				requestMethod:  http.MethodPost,
				requestString:  "value",
				parametersJson: []byte(`{"id":"PollCounter", "type":"counter"}`),
				containsValue:  "",
				msgString:      "POST value",
				mock: func(m *mocks.MockStorage) {
					m.EXPECT().
						GetValue(gomock.Any(), "counter", "PollCounter").
						Return(
							nil,
							errors.New("not found"),
						).
						Times(1)
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	if err := logger.Initialize("info"); err != nil {
		t.Fatal(err.Error())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.requestMethod, "/"+tt.want.requestString, bytes.NewBuffer(tt.want.parametersJson))

			tt.want.mock(mockStorage)

			w := httptest.NewRecorder()

			GetMetricValueByJsonHandler(t.Context(), w, request, mockStorage)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode, tt.want.msgString)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if res.Body != nil {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("can't close response body: %s", err.Error())
				}
			}
			assert.Contains(t, string(resBody), tt.want.containsValue)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestPostgresPingHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive ping",
			args: args{
				w: httptest.NewRecorder(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	m.EXPECT().Ping(gomock.Any()).Return(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PingHandler(t.Context(), tt.args.w, nil, m)
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(`[]`))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	m.EXPECT().SetValues(gomock.Any(), []models.Metrics{}).Return(nil)

	SetMetrics(t.Context(), w, r, m)

	assert.Equal(t, http.StatusOK, w.Code)
}
