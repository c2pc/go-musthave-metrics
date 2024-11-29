package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/config"
	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
)

const requestTimeout = 1 * time.Second

type Hasher interface {
	Hash([]byte) (string, error)
}

type Client struct {
	serverAddr string
	hasher     Hasher
}

func NewClient(serverAddr string, hasher Hasher) reporter.Updater {
	if !strings.Contains(serverAddr, "http") {
		serverAddr = "http://" + serverAddr
	}

	return &Client{
		serverAddr: serverAddr,
		hasher:     hasher,
	}
}

func (c *Client) UpdateMetric(ctx context.Context, metrics []model.Metrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	// Сжимаем данные в формате gzip
	buf, err := c.compress(body)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: requestTimeout,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverAddr+"/updates/", buf)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	if c.hasher != nil {
		hash, err := c.hash(body)
		if err != nil {
			return err
		}
		request.Header.Set(config.HashHeader, hash)
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

func (c *Client) compress(data []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (c *Client) hash(data []byte) (string, error) {
	hash, err := c.hasher.Hash(data)
	if err != nil {
		return "", err
	}

	return hash, nil
}
