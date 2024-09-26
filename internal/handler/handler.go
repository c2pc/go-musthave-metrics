package handler

import (
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"net/http"
)

func NewHandler(gaugeStorage storage.Storage[float64], counterStorage storage.Storage[int64]) http.Handler {
	mux := http.NewServeMux()

	newMetricUpdateHandler(gaugeStorage, counterStorage).init(mux)

	return mux
}
