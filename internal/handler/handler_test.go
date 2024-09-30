package handler_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		gaugeStorage   storage.Storage[float64]
		counterStorage storage.Storage[int64]
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty storages",
			args: args{
				nil, nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "empty gauge storage",
			args: args{
				nil, storage.NewCounterStorage(),
			},
			wantErr: assert.Error,
		},
		{
			name: "empty counter storage",
			args: args{
				storage.NewGaugeStorage(), nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "success",
			args: args{
				storage.NewGaugeStorage(), storage.NewCounterStorage(),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.NewHandler(tt.args.gaugeStorage, tt.args.counterStorage)
			if !tt.wantErr(t, err, fmt.Sprintf("NewHandler(%v, %v)", tt.args.gaugeStorage, tt.args.counterStorage)) {
				return
			}
		})
	}
}

func TestMetricHandler_HandleUpdate(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handler2, err := handler.NewHandler(gaugeStorage, counterStorage)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{"Get", http.MethodGet, "/update/gauge/metric/5", http.StatusNotFound},
		{"Put", http.MethodPut, "/update/counter/metric/5", http.StatusNotFound},
		{"Patch", http.MethodPatch, "/update/gauge/metric/5", http.StatusNotFound},
		{"Delete", http.MethodDelete, "/update/counter/metric/5", http.StatusNotFound},
		{"Connect", http.MethodConnect, "/update/gauge/metric/5", http.StatusNotFound},
		{"Options", http.MethodOptions, "/update/counter/metric/5", http.StatusNotFound},
		{"Trace", http.MethodTrace, "/update/gauge/metric/5", http.StatusNotFound},
		{"Head", http.MethodHead, "/update/counter/metric/5", http.StatusNotFound},

		{"Empty type", http.MethodPost, "/update//metric/4", http.StatusBadRequest},
		{"Empty name", http.MethodPost, "/update/gauge/5", http.StatusNotFound},
		{"Empty name2", http.MethodPost, "/update/counter/7", http.StatusNotFound},
		{"Empty name3", http.MethodPost, "/update/gauge/8", http.StatusNotFound},
		{"Empty value", http.MethodPost, "/update/counter/metric", http.StatusNotFound},
		{"Empty value2", http.MethodPost, "/update/gauge/metric", http.StatusNotFound},

		{"Invalid type", http.MethodPost, "/update/invalid/metric/12", http.StatusBadRequest},
		{"Invalid value", http.MethodPost, "/update/gauge/metric/invalid", http.StatusBadRequest},
		{"Invalid value2", http.MethodPost, "/update/counter/metric/invalid", http.StatusBadRequest},

		{"Success Gauge", http.MethodPost, "/update/gauge/metric/1", http.StatusOK},
		{"Success Counter", http.MethodPost, "/update/counter/metric/1", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)

			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestMetricHandler_HandleValue(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handler2, err := handler.NewHandler(gaugeStorage, counterStorage)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{"Post", http.MethodPost, "/value/gauge/metric", http.StatusNotFound, ""},
		{"Put", http.MethodPut, "/value/counter/metric", http.StatusNotFound, ""},
		{"Patch", http.MethodPatch, "/value/gauge/metric", http.StatusNotFound, ""},
		{"Delete", http.MethodDelete, "/value/counter/metric", http.StatusNotFound, ""},
		{"Connect", http.MethodConnect, "/value/gauge/metric", http.StatusNotFound, ""},
		{"Options", http.MethodOptions, "/value/counter/metric", http.StatusNotFound, ""},
		{"Trace", http.MethodTrace, "/value/gauge/metric", http.StatusNotFound, ""},
		{"Head", http.MethodHead, "/value/counter/metric", http.StatusNotFound, ""},

		{"Empty type", http.MethodGet, "/value//metric", http.StatusBadRequest, ""},
		{"Empty name", http.MethodGet, "/value/gauge", http.StatusNotFound, ""},
		{"Empty name2", http.MethodGet, "/value/counter", http.StatusNotFound, ""},

		{"Invalid type", http.MethodGet, "/value/invalid/metric", http.StatusBadRequest, ""},

		{"Not found Gauge", http.MethodGet, "/value/gauge/metric1", http.StatusNotFound, ""},
		{"Not found Counter", http.MethodGet, "/value/counter/metric1", http.StatusNotFound, ""},

		{"Success Gauge", http.MethodGet, "/value/gauge/metric2", http.StatusOK, "20"},
		{"Success Counter", http.MethodGet, "/value/counter/metric2", http.StatusOK, "30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedBody != "" {
				split := strings.Split(tt.url, "/")
				if len(split) < 4 {
					t.Errorf("Invalid URL: %s", tt.url)
				}

				if split[2] == "gauge" {
					value, err := strconv.ParseFloat(tt.expectedBody, 64)
					require.NoError(t, err)

					err = gaugeStorage.Set(split[3], value)
					require.NoError(t, err)
				} else if split[2] == "counter" {
					value, err := strconv.ParseInt(tt.expectedBody, 10, 64)
					require.NoError(t, err)

					err = counterStorage.Set(split[3], value)
					require.NoError(t, err)
				}
			}

			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if tt.expectedBody != "" {
				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				split := strings.Split(tt.url, "/")
				if len(split) < 4 {
					t.Errorf("Invalid URL: %s", tt.url)
				}

				if split[2] == "gauge" {
					got, err := strconv.ParseFloat(tt.expectedBody, 64)
					require.NoError(t, err)

					rValue, err := strconv.ParseFloat(string(body), 64)
					require.NoError(t, err)

					assert.Equal(t, got, rValue)
				} else if split[2] == "counter" {
					got, err := strconv.ParseInt(tt.expectedBody, 10, 64)
					require.NoError(t, err)

					rValue, err := strconv.ParseInt(string(body), 10, 64)
					require.NoError(t, err)

					assert.Equal(t, got, rValue)
				}

			}

			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestMetricHandler_HandleAll(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handler2, err := handler.NewHandler(gaugeStorage, counterStorage)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		data           []string
		expectedStatus int
	}{
		{"Post", http.MethodPost, nil, http.StatusNotFound},
		{"Put", http.MethodPut, nil, http.StatusNotFound},
		{"Patch", http.MethodPatch, nil, http.StatusNotFound},
		{"Delete", http.MethodDelete, nil, http.StatusNotFound},
		{"Connect", http.MethodConnect, nil, http.StatusNotFound},
		{"Options", http.MethodOptions, nil, http.StatusNotFound},
		{"Trace", http.MethodTrace, nil, http.StatusNotFound},
		{"Head", http.MethodHead, nil, http.StatusNotFound},

		{"Success", http.MethodGet, []string{"gauge/metric/1"}, http.StatusOK},
		{"Success2", http.MethodGet, []string{"counter/metric/1"}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data != nil {
				for _, data := range tt.data {
					split := strings.Split(data, "/")
					if len(split) < 3 {
						t.Errorf("Invalid data: %s", data)
					}

					if split[0] == "gauge" {
						value, err := strconv.ParseFloat(split[2], 64)
						require.NoError(t, err)

						err = gaugeStorage.Set(split[1], value)
						require.NoError(t, err)
					} else if split[2] == "counter" {
						value, err := strconv.ParseInt(split[2], 10, 64)
						require.NoError(t, err)

						err = counterStorage.Set(split[1], value)
						require.NoError(t, err)
					}
				}

			}

			request := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, result.Header.Get("Content-Type"), "text/html; charset=utf-8")
			}

			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}
