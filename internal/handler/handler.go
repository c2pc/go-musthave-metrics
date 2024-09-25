package handler

import (
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"net/http"
)

func NewHandler(gaugeStorage storage.Storage[storage.GaugeMetricValue], counterStorage storage.Storage[storage.CounterMetricValue]) http.Handler {
	mux := http.NewServeMux()

	newMetricUpdateHandler(gaugeStorage, counterStorage).init(mux)

	return mux
}
