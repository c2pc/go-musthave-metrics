package handler

import (
	"fmt"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"net/http"
)

func NewHandler(gaugeStorage storage.Storage[float64], counterStorage storage.Storage[int64]) (http.Handler, error) {
	if counterStorage == nil {
		return nil, fmt.Errorf("counterStorage is nil")
	}

	if gaugeStorage == nil {
		return nil, fmt.Errorf("gaugeStorage is nil")
	}

	mux := http.NewServeMux()

	NewMetricUpdateHandler(gaugeStorage, counterStorage).Init(mux)

	return mux, nil
}
