package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
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

	// Сжимаем данные в формате gzip
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(body); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverAddr+"/update/", &buf)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var respBody []byte
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		respBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
	} else {
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	var metricRes model.Metrics
	if err := json.Unmarshal(respBody, &metricRes); err != nil {
		return nil, err
	}

	return &metricRes, nil
}
