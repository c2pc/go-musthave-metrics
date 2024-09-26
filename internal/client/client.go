package client

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
)

type IClient interface {
	UpdateMetric(tp string, name string, value interface{}) error
}

type Client struct {
	serverAddr string
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) UpdateMetric(tp string, name string, value interface{}) error {
	var url = "/update/" + tp + "/" + name

	switch val := value.(type) {
	case int64:
		url += "/" + strconv.FormatInt(val, 10)
	case float64:
		url += "/" + strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return errors.New("invalid metric value type")
	}

	client := &http.Client{}
	var body []byte
	request, err := http.NewRequest(http.MethodPost, c.serverAddr+url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
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
