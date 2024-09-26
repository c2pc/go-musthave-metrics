package handler

import (
	"fmt"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"net/http"
	"strconv"
)

type MetricUpdateHandler struct {
	gaugeStorage   storage.Storage[float64]
	counterStorage storage.Storage[int64]
}

func newMetricUpdateHandler(
	gaugeStorage storage.Storage[float64],
	counterStorage storage.Storage[int64],
) *MetricUpdateHandler {

	return &MetricUpdateHandler{gaugeStorage, counterStorage}
}

func (h *MetricUpdateHandler) init(mux *http.ServeMux) {
	mux.HandleFunc("/update/{type}/{name}/{value}", h.handle)
}

func (h *MetricUpdateHandler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metricType string
	if metricType = r.PathValue("type"); metricType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = r.PathValue("name"); metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var metricValue string
	if metricValue = r.PathValue("value"); metricValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err error

	fmt.Println("Updating ", metricType, metricName, metricValue)
	switch metricType {
	case h.gaugeStorage.GetName():
		var value float64
		if value, err = strconv.ParseFloat(metricValue, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := h.gaugeStorage.Set(metricName, value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case h.counterStorage.GetName():
		var value int64
		if value, err = strconv.ParseInt(metricValue, 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := h.counterStorage.Set(metricName, value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
