package handler_test

import (
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserViewHandler(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	updateHandler := handler.NewMetricUpdateHandler(gaugeStorage, counterStorage)

	tests := []struct {
		name           string
		method         string
		metricType     string
		metricName     string
		metricValue    string
		expectedStatus int
	}{
		{"Get", http.MethodGet, "gauge", "metric", "5", http.StatusMethodNotAllowed},
		{"Put", http.MethodPut, "counter", "metric", "5", http.StatusMethodNotAllowed},
		{"Patch", http.MethodPatch, "gauge", "metric", "5", http.StatusMethodNotAllowed},
		{"Delete", http.MethodDelete, "counter", "metric", "5", http.StatusMethodNotAllowed},
		{"Connect", http.MethodConnect, "gauge", "metric", "5", http.StatusMethodNotAllowed},
		{"Options", http.MethodOptions, "counter", "metric", "5", http.StatusMethodNotAllowed},
		{"Trace", http.MethodTrace, "gauge", "metric", "5", http.StatusMethodNotAllowed},
		{"Head", http.MethodHead, "counter", "metric", "5", http.StatusMethodNotAllowed},

		{"Empty type", http.MethodPost, "", "metric", "4", http.StatusBadRequest},
		{"Empty name", http.MethodPost, "gauge", "", "5", http.StatusNotFound},
		{"Empty name2", http.MethodPost, "counter", "", "7", http.StatusNotFound},
		{"Empty name3", http.MethodPost, "gauge", "", "8", http.StatusNotFound},
		{"Empty value", http.MethodPost, "counter", "metric", "", http.StatusBadRequest},
		{"Empty value2", http.MethodPost, "gauge", "metric", "", http.StatusBadRequest},

		{"Invalid type", http.MethodPost, "invalid", "metric", "12", http.StatusBadRequest},
		{"Invalid value", http.MethodPost, "gauge", "metric", "invalid", http.StatusBadRequest},
		{"Invalid value2", http.MethodPost, "counter", "metric", "invalid", http.StatusBadRequest},

		{"Success Gauge", http.MethodPost, "gauge", "metric", "1", http.StatusOK},
		{"Success Counter", http.MethodPost, "counter", "metric", "1", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/update"+tt.metricType+"/"+tt.metricName+"/"+tt.metricValue, nil)
			request.SetPathValue("type", tt.metricType)
			request.SetPathValue("name", tt.metricName)
			request.SetPathValue("value", tt.metricValue)
			w := httptest.NewRecorder()

			h := http.HandlerFunc(updateHandler.HandleUpdate)
			h(w, request)

			result := w.Result()
			assert.Equal(t, tt.expectedStatus, result.StatusCode)

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
