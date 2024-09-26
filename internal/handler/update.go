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
		fmt.Println("metric type is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = r.PathValue("name"); metricName == "" {
		fmt.Println("metric name is empty")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var metricValue string
	if metricValue = r.PathValue("value"); metricValue == "" {
		fmt.Println("metric value is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err error

	fmt.Println("Request ", metricType, metricName, metricValue)

	switch metricType {
	case h.gaugeStorage.GetName():
		var value float64
		if value, err = strconv.ParseFloat(metricValue, 64); err != nil {
			fmt.Println("could not parse value to float64", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := h.gaugeStorage.Set(metricName, value); err != nil {
			fmt.Println("could not store metric value", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case h.counterStorage.GetName():
		var value int64
		if value, err = strconv.ParseInt(metricValue, 10, 64); err != nil {
			fmt.Println("could not parse value to int64", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := h.counterStorage.Set(metricName, value); err != nil {
			fmt.Println("could not store metric value", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		fmt.Println("unknown metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("Success ", metricType, metricName, metricValue)

	w.WriteHeader(http.StatusOK)
}
