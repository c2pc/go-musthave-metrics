package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
)

type Client struct {
	serverAddr string
}

func NewClient(serverAddr string) reporter.Updater {
	if !strings.Contains(serverAddr, "http") {
		serverAddr = "http://" + serverAddr
	}

	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) UpdateMetric(ctx context.Context, tp string, name string, value interface{}) (*model.Metrics, error) {
	metricRequest := model.Metrics{
		ID:    name,
		MType: tp,
	}

	switch val := value.(type) {
	case int64:
		metricRequest.Delta = &val
	case float64:
		metricRequest.Value = &val
	default:
		return nil, errors.New("invalid metric value type")
	}

	body, err := json.Marshal(metricRequest)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverAddr+"/update/", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var metricRes model.Metrics
	if err := json.NewDecoder(response.Body).Decode(&metricRes); err != nil {
		return nil, err
	}

	return &metricRes, nil
}
