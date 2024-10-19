package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

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

func (c *Client) UpdateMetric(ctx context.Context, tp string, name string, value interface{}) error {
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
		return errors.New("invalid metric value type")
	}

	body, err := json.Marshal(metricRequest)
	if err != nil {
		return err
	}

	client := &http.Client{}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverAddr+"/update/", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return err
	}

	return nil
}
