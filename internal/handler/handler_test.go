package handler_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/c2pc/go-musthave-metrics/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
)

func TestMetricHandler_HandleUpdate(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
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

func TestMetricHandler_HandleUpdateJSON(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	var defaultDelta, defaultDelta2 int64 = 10, -6
	var defaultValue, defaultValue2 float64 = 10, 0.5

	type data struct {
		ID    interface{} `json:"id"`
		Type  interface{} `json:"type"`
		Delta interface{} `json:"delta,omitempty"`
		Value interface{} `json:"value,omitempty"`
	}

	tests := []struct {
		name           string
		method         string
		data           *data
		expectedStatus int
	}{
		{"Get", http.MethodGet, nil, http.StatusNotFound},
		{"Put", http.MethodPut, nil, http.StatusNotFound},
		{"Patch", http.MethodPatch, nil, http.StatusNotFound},
		{"Delete", http.MethodDelete, nil, http.StatusNotFound},
		{"Connect", http.MethodConnect, nil, http.StatusNotFound},
		{"Options", http.MethodOptions, nil, http.StatusNotFound},
		{"Trace", http.MethodTrace, nil, http.StatusNotFound},
		{"Head", http.MethodHead, nil, http.StatusNotFound},

		{"Empty body", http.MethodPost, nil, http.StatusBadRequest},
		{"Empty type", http.MethodPost, &data{ID: "id", Type: "", Delta: defaultDelta, Value: defaultValue}, http.StatusBadRequest},
		{"Empty name", http.MethodPost, &data{ID: "", Type: "gauge", Delta: defaultDelta2, Value: defaultValue2}, http.StatusNotFound},
		{"Empty name2", http.MethodPost, &data{ID: "", Type: "counter", Delta: defaultDelta, Value: defaultValue2}, http.StatusNotFound},
		{"Empty name3", http.MethodPost, &data{ID: "", Type: "gauge", Delta: defaultDelta2, Value: defaultValue}, http.StatusNotFound},
		{"Empty value", http.MethodPost, &data{ID: "id1", Type: "counter", Delta: nil, Value: nil}, http.StatusBadRequest},
		{"Empty value2", http.MethodPost, &data{ID: "id2", Type: "gauge", Delta: nil, Value: nil}, http.StatusBadRequest},

		{"Invalid type", http.MethodPost, &data{ID: "id3", Type: "invalid", Delta: defaultDelta2, Value: defaultValue2}, http.StatusBadRequest},
		{"Invalid value", http.MethodPost, &data{ID: "id3", Type: "gauge", Value: "invalid"}, http.StatusBadRequest},
		{"Invalid value2", http.MethodPost, &data{ID: "id3", Type: "counter", Delta: "invalid"}, http.StatusBadRequest},

		{"Success Gauge", http.MethodPost, &data{ID: "id2", Type: "gauge", Value: defaultValue2}, http.StatusOK},
		{"Success Counter", http.MethodPost, &data{ID: "id3", Type: "counter", Delta: defaultDelta2}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.data != nil {
				out, err := json.Marshal(tt.data)
				if err != nil {
					log.Fatal(err)
				}
				body = bytes.NewReader(out)
			}

			request := httptest.NewRequest(tt.method, "/update/", body)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			err := result.Body.Close()
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				res, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				var metrics model.Metrics
				err = json.Unmarshal(res, &metrics)
				require.NoError(t, err)

				assert.Equal(t, tt.data.ID, metrics.ID)
				assert.Equal(t, tt.data.Type, metrics.Type)

				switch tt.data.Type {
				case gaugeStorage.GetName():
					value, err := gaugeStorage.Get(context.Background(), tt.data.ID.(string))
					require.NoError(t, err)
					assert.Equal(t, tt.data.Value, value)
					assert.NotEmpty(t, *metrics.Value)
					assert.Equal(t, tt.data.Value, *metrics.Value)
				case counterStorage.GetName():
					value, err := counterStorage.Get(context.Background(), tt.data.ID.(string))
					require.NoError(t, err)
					assert.Equal(t, tt.data.Delta, value)
					assert.NotEmpty(t, *metrics.Delta)
					assert.Equal(t, tt.data.Delta, *metrics.Delta)

				default:
					assert.Fail(t, "unknown metrics type")
				}
			}

		})
	}
}

func TestMetricHandler_HandleUpdatesJSON(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	var defaultDelta, defaultDelta2 int64 = 10, -6
	var defaultValue, defaultValue2 float64 = 10, 0.5

	type data struct {
		ID    interface{} `json:"id"`
		Type  interface{} `json:"type"`
		Delta interface{} `json:"delta,omitempty"`
		Value interface{} `json:"value,omitempty"`
	}

	tests := []struct {
		name           string
		method         string
		data           []data
		expectedStatus int
	}{
		{"Get", http.MethodGet, nil, http.StatusNotFound},
		{"Put", http.MethodPut, nil, http.StatusNotFound},
		{"Patch", http.MethodPatch, nil, http.StatusNotFound},
		{"Delete", http.MethodDelete, nil, http.StatusNotFound},
		{"Connect", http.MethodConnect, nil, http.StatusNotFound},
		{"Options", http.MethodOptions, nil, http.StatusNotFound},
		{"Trace", http.MethodTrace, nil, http.StatusNotFound},
		{"Head", http.MethodHead, nil, http.StatusNotFound},

		{"Empty body", http.MethodPost, nil, http.StatusBadRequest},
		{"Empty type", http.MethodPost, []data{{ID: "id", Type: "", Delta: defaultDelta, Value: defaultValue}}, http.StatusBadRequest},
		{"Empty name", http.MethodPost, []data{{ID: "", Type: "gauge", Delta: defaultDelta2, Value: defaultValue2}}, http.StatusBadRequest},
		{"Empty name2", http.MethodPost, []data{{ID: "", Type: "counter", Delta: defaultDelta, Value: defaultValue2}}, http.StatusBadRequest},
		{"Empty name3", http.MethodPost, []data{{ID: "", Type: "gauge", Delta: defaultDelta2, Value: defaultValue}}, http.StatusBadRequest},
		{"Empty value", http.MethodPost, []data{{ID: "id1", Type: "counter", Delta: nil, Value: nil}}, http.StatusBadRequest},
		{"Empty value2", http.MethodPost, []data{{ID: "id2", Type: "gauge", Delta: nil, Value: nil}}, http.StatusBadRequest},

		{"Invalid type", http.MethodPost, []data{{ID: "id3", Type: "invalid", Delta: defaultDelta2, Value: defaultValue2}}, http.StatusBadRequest},
		{"Invalid value", http.MethodPost, []data{{ID: "id3", Type: "gauge", Value: "invalid"}}, http.StatusBadRequest},
		{"Invalid value2", http.MethodPost, []data{{ID: "id3", Type: "counter", Delta: "invalid"}}, http.StatusBadRequest},

		{"Success Gauge", http.MethodPost, []data{{ID: "id2", Type: "gauge", Value: defaultValue2}}, http.StatusOK},
		{"Success Counter", http.MethodPost, []data{{ID: "id3", Type: "counter", Delta: defaultDelta2}}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.data != nil {
				out, err := json.Marshal(tt.data)
				if err != nil {
					log.Fatal(err)
				}
				body = bytes.NewReader(out)
			}

			request := httptest.NewRequest(tt.method, "/updates/", body)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)
			result.Body.Close()
		})
	}
}

func TestMetricHandler_HandleUpdateJSON_Compress(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	tests := []struct {
		name           string
		value          float64
		compress       bool
		decompress     bool
		expectedStatus int
	}{
		{"Compress", 1, true, false, http.StatusOK},
		{"Decompress", 2, false, true, http.StatusOK},
		{"All", 3, true, true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data = struct {
				ID    string  `json:"id"`
				Type  string  `json:"type"`
				Delta int64   `json:"delta,omitempty"`
				Value float64 `json:"value,omitempty"`
			}{
				ID:    "1",
				Type:  gaugeStorage.GetName(),
				Delta: 1,
				Value: tt.value,
			}

			buf := new(bytes.Buffer)

			out, err := json.Marshal(data)
			require.NoError(t, err)

			if tt.compress {
				zb := gzip.NewWriter(buf)
				defer zb.Close()

				_, err = zb.Write(out)
				require.NoError(t, err)

				err = zb.Close()
				require.NoError(t, err)
			} else {
				buf.Write(out)
			}

			request := httptest.NewRequest(http.MethodPost, "/update/", buf)
			request.Header.Set("Content-Type", "application/json")
			if tt.compress {
				request.Header.Set("Content-Encoding", "gzip")
			}
			if tt.decompress {
				request.Header.Set("Accept-Encoding", "gzip")
			}

			w := httptest.NewRecorder()
			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if result.StatusCode == http.StatusOK {
				var res []byte
				if tt.decompress {
					zr, err := gzip.NewReader(result.Body)
					require.NoError(t, err)
					defer zr.Close() // Обязательно закрываем gzip.Reader

					res, err = io.ReadAll(zr)
					require.NoError(t, err)
				} else {
					res, err = io.ReadAll(result.Body)
					require.NoError(t, err)
				}

				var metrics model.Metrics
				err = json.Unmarshal(res, &metrics)
				require.NoError(t, err)

				assert.Equal(t, data.ID, metrics.ID)
				assert.Equal(t, data.Type, metrics.Type)

				value, err := gaugeStorage.Get(context.Background(), data.ID)
				require.NoError(t, err)
				assert.Equal(t, data.Value, value)
				assert.NotEmpty(t, *metrics.Value)
				assert.Equal(t, data.Value, *metrics.Value)
			}

			require.NoError(t, result.Body.Close())
		})
	}
}

func TestMetricHandler_HandleValue(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
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

					err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: split[3], Value: value})
					require.NoError(t, err)
				} else if split[2] == "counter" {
					value, err := strconv.ParseInt(tt.expectedBody, 10, 64)
					require.NoError(t, err)

					err = counterStorage.Set(context.Background(), storage.Value[int64]{Key: split[3], Value: value})
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

func TestMetricHandler_HandleValueJSON(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	var defaultDelta int64 = -6
	var defaultValue = 0.5

	type data struct {
		ID   interface{} `json:"id"`
		Type interface{} `json:"type"`
	}
	type expectedData struct {
		ID    interface{} `json:"id"`
		Type  interface{} `json:"type"`
		Delta interface{} `json:"delta,omitempty"`
		Value interface{} `json:"value,omitempty"`
	}

	tests := []struct {
		name           string
		method         string
		data           *data
		expectedStatus int
		expectedBody   *expectedData
	}{
		{"Post", http.MethodGet, nil, http.StatusNotFound, nil},
		{"Put", http.MethodPut, nil, http.StatusNotFound, nil},
		{"Patch", http.MethodPatch, nil, http.StatusNotFound, nil},
		{"Delete", http.MethodDelete, nil, http.StatusNotFound, nil},
		{"Connect", http.MethodConnect, nil, http.StatusNotFound, nil},
		{"Options", http.MethodOptions, nil, http.StatusNotFound, nil},
		{"Trace", http.MethodTrace, nil, http.StatusNotFound, nil},
		{"Head", http.MethodHead, nil, http.StatusNotFound, nil},

		{"Empty body", http.MethodPost, nil, http.StatusBadRequest, nil},
		{"Empty type", http.MethodPost, &data{ID: "id", Type: ""}, http.StatusBadRequest, nil},
		{"Empty name", http.MethodPost, &data{ID: "", Type: "gauge"}, http.StatusNotFound, nil},
		{"Empty name2", http.MethodPost, &data{ID: "", Type: "gauge"}, http.StatusNotFound, nil},

		{"Invalid type", http.MethodPost, &data{ID: "id3", Type: "invalid"}, http.StatusBadRequest, nil},

		{"Not found Gauge", http.MethodPost, &data{ID: "id4", Type: "gauge"}, http.StatusNotFound, nil},
		{"Not found Counter", http.MethodPost, &data{ID: "id5", Type: "counter"}, http.StatusNotFound, nil},

		{"Success Gauge", http.MethodPost, &data{ID: "id41", Type: "gauge"}, http.StatusOK, &expectedData{ID: "id41", Type: "gauge", Value: defaultValue}},
		{"Success Counter", http.MethodPost, &data{ID: "id42", Type: "counter"}, http.StatusOK, &expectedData{ID: "id42", Type: "counter", Delta: defaultDelta}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedBody != nil {
				if tt.data.Type == "gauge" {
					err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.data.ID.(string), Value: tt.expectedBody.Value.(float64)})
					require.NoError(t, err)
				} else if tt.data.Type == "counter" {
					err = counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.data.ID.(string), Value: tt.expectedBody.Delta.(int64)})
					require.NoError(t, err)
				}
			}

			var body io.Reader
			if tt.data != nil {
				out, err := json.Marshal(tt.data)
				if err != nil {
					log.Fatal(err)
				}
				body = bytes.NewReader(out)
			}

			request := httptest.NewRequest(tt.method, "/value/", body)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if tt.expectedBody != nil {
				res, err := io.ReadAll(result.Body)
				require.NoError(t, err)

				var metrics model.Metrics
				err = json.Unmarshal(res, &metrics)
				require.NoError(t, err)

				assert.Equal(t, tt.data.ID, metrics.ID)
				assert.Equal(t, tt.data.Type, metrics.Type)

				switch tt.data.Type {
				case gaugeStorage.GetName():
					value, err := gaugeStorage.Get(context.Background(), tt.data.ID.(string))
					require.NoError(t, err)
					assert.Equal(t, tt.expectedBody.Value, value)
					assert.NotEmpty(t, *metrics.Value)
					assert.Equal(t, tt.expectedBody.Value, *metrics.Value)
				case counterStorage.GetName():
					value, err := counterStorage.Get(context.Background(), tt.data.ID.(string))
					require.NoError(t, err)
					assert.Equal(t, tt.expectedBody.Delta, value)
					assert.NotEmpty(t, *metrics.Delta)
					assert.Equal(t, tt.expectedBody.Delta, *metrics.Delta)

				default:
					assert.Fail(t, "unknown metrics type")
				}
			}

			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestMetricHandler_HandleValueJSON_Compress(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	tests := []struct {
		name           string
		value          float64
		compress       bool
		decompress     bool
		expectedStatus int
	}{
		{"Compress", 1, true, false, http.StatusOK},
		{"Decompress", 2, false, true, http.StatusOK},
		{"All", 3, true, true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data = struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			}{
				ID:   "1",
				Type: gaugeStorage.GetName(),
			}

			buf := new(bytes.Buffer)

			err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: data.ID, Value: tt.value})
			require.NoError(t, err)

			out, err := json.Marshal(data)
			require.NoError(t, err)

			if tt.compress {
				zb := gzip.NewWriter(buf)
				defer zb.Close()

				_, err = zb.Write(out)
				require.NoError(t, err)

				err = zb.Close()
				require.NoError(t, err)
			} else {
				buf.Write(out)
			}

			request := httptest.NewRequest(http.MethodPost, "/value/", buf)
			request.Header.Set("Content-Type", "application/json")
			if tt.compress {
				request.Header.Set("Content-Encoding", "gzip")
			}
			if tt.decompress {
				request.Header.Set("Accept-Encoding", "gzip")
			}

			w := httptest.NewRecorder()
			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if result.StatusCode == http.StatusOK {
				var res []byte
				if tt.decompress {
					zr, err := gzip.NewReader(result.Body)
					require.NoError(t, err)
					defer zr.Close() // Обязательно закрываем gzip.Reader

					res, err = io.ReadAll(zr)
					require.NoError(t, err)
				} else {
					res, err = io.ReadAll(result.Body)
					require.NoError(t, err)
				}

				var metrics model.Metrics
				err = json.Unmarshal(res, &metrics)
				require.NoError(t, err)

				assert.Equal(t, data.ID, metrics.ID)
				assert.Equal(t, data.Type, metrics.Type)

				value, err := gaugeStorage.Get(context.Background(), data.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.value, value)
				assert.NotEmpty(t, *metrics.Value)
				assert.Equal(t, tt.value, *metrics.Value)
			}

			require.NoError(t, result.Body.Close())
		})
	}
}
func TestMetricHandler_HandleAll(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
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

						err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: split[1], Value: value})
						require.NoError(t, err)
					} else if split[2] == "counter" {
						value, err := strconv.ParseInt(split[2], 10, 64)
						require.NoError(t, err)

						err = counterStorage.Set(context.Background(), storage.Value[int64]{Key: split[1], Value: value})
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

func TestMetricHandler_HandleAll_Compress(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	m, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, m)
	require.NoError(t, err)

	tests := []struct {
		name           string
		decompress     bool
		expectedStatus int
	}{
		{"No Decompress", false, http.StatusOK},
		{"Decompress", true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: "id", Value: 10})
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.decompress {
				request.Header.Set("Accept-Encoding", "gzip")
			}

			w := httptest.NewRecorder()
			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			if result.StatusCode == http.StatusOK {
				if tt.decompress {
					zr, err := gzip.NewReader(result.Body)
					require.NoError(t, err)
					defer zr.Close() // Обязательно закрываем gzip.Reader

					_, err = io.ReadAll(zr)
					require.NoError(t, err)
				}
			}

			require.NoError(t, result.Body.Close())
		})
	}
}

func TestMetricHandler_Ping(t *testing.T) {
	m, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer m.Close()

	db := database.DB{DB: m}

	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	handler2 := handler.NewHandler(gaugeStorage, counterStorage, &db)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		mockgen        func()
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

		{"No Ping", http.MethodGet, func() {
			mock.ExpectPing().WillReturnError(fmt.Errorf("ping error"))
		}, http.StatusInternalServerError},

		{"Success", http.MethodGet, func() {
			mock.ExpectPing().WillReturnError(nil)
		}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockgen != nil {
				tt.mockgen()
			}
			request := httptest.NewRequest(tt.method, "/ping", nil)

			w := httptest.NewRecorder()

			handler2.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			require.NoError(t, result.Body.Close())
		})
	}
}
