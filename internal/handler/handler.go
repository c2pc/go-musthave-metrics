package handler

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/retry"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/gin-gonic/gin"

	"github.com/c2pc/go-musthave-metrics/internal/handler/middleware"
	"github.com/c2pc/go-musthave-metrics/internal/model"
)

type Storager[T int64 | float64] interface {
	GetName() string
	Get(ctx context.Context, key string) (T, error)
	GetString(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context) (map[string]T, error)
	GetAllString(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, values ...storage.Valuer[T]) error
	SetString(ctx context.Context, values ...storage.Valuer[string]) error
}

type Pinger interface {
	Ping() error
}

type Handler struct {
	http.Handler
	gaugeStorage   Storager[float64]
	counterStorage Storager[int64]
	db             Pinger
}

func NewHandler(gaugeStorage Storager[float64], counterStorage Storager[int64], db Pinger) http.Handler {

	gin.SetMode(gin.ReleaseMode)
	handlers := gin.New()

	handlers.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	h := &Handler{
		Handler:        handlers,
		gaugeStorage:   gaugeStorage,
		counterStorage: counterStorage,
		db:             db,
	}

	h.Init(handlers)

	return h
}

func (h *Handler) Init(engine *gin.Engine) {
	api := engine.Group("", middleware.GzipDecompressor, middleware.GzipCompressor, middleware.Logger)
	{
		api.GET("/", h.handleHTML)
		api.GET("/ping", h.ping)
		api.POST("/update/", h.handleUpdateJSON)
		api.POST("/updates/", h.handleUpdatesJSON)
		api.POST("/update/:type/:name/:value", h.handleUpdate)
		api.GET("/value/:type/:name", h.handleValue)
		api.POST("/value/", h.handleValueJSON)
	}
}

func (h *Handler) handleUpdate(c *gin.Context) {
	ctx := c.Request.Context()

	var metricType string
	if metricType = c.Param("type"); metricType == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = c.Param("name"); metricName == "" {
		c.Status(http.StatusNotFound)
		return
	}

	var metricValue string
	if metricValue = c.Param("value"); metricValue == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	switch metricType {
	case h.gaugeStorage.GetName():
		if err := h.gaugeStorage.SetString(ctx, storage.Value[string]{Key: metricName, Value: metricValue}); err != nil {
			if errors.Is(err, storage.ErrInvalidValue) {
				c.Status(http.StatusBadRequest)
				return
			}

			c.Status(http.StatusInternalServerError)
			return
		}

	case h.counterStorage.GetName():
		if err := h.counterStorage.SetString(ctx, storage.Value[string]{Key: metricName, Value: metricValue}); err != nil {
			if errors.Is(err, storage.ErrInvalidValue) {
				c.Status(http.StatusBadRequest)
				return
			}

			c.Status(http.StatusInternalServerError)
			return
		}

	default:
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleUpdateJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var metric model.Metrics
	message, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	err = json.Unmarshal(message, &metric)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal request body"})
		return
	}

	if metric.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The metric type is empty"})
		return
	}

	if metric.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "The metric id is empty"})
		return
	}

	var metricRequest *model.Metrics

	switch metric.Type {
	case h.gaugeStorage.GetName():
		if metric.Value == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The metric value is empty"})
			return
		}

		if err := retry.Retry(
			func() error {
				return h.gaugeStorage.Set(ctx, storage.Value[float64]{Key: metric.ID, Value: *metric.Value})
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrInvalidValue) {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set metric value"})
			return
		}

		newValue, err := h.gaugeStorage.Get(ctx, metric.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metric value"})
			return
		}

		metricRequest = &model.Metrics{
			Type:  h.gaugeStorage.GetName(),
			ID:    metric.ID,
			Value: &newValue,
		}

	case h.counterStorage.GetName():
		if metric.Delta == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The metric delta is empty"})
			return
		}

		if err := retry.Retry(
			func() error {
				return h.counterStorage.Set(ctx, storage.Value[int64]{Key: metric.ID, Value: *metric.Delta})
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrInvalidValue) {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set metric value"})
			return
		}

		newValue, err := h.counterStorage.Get(ctx, metric.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metric value"})
			return
		}

		metricRequest = &model.Metrics{
			Type:  h.counterStorage.GetName(),
			ID:    metric.ID,
			Delta: &newValue,
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metrics type"})
		return
	}

	c.JSON(http.StatusOK, metricRequest)
}

func (h *Handler) handleUpdatesJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var metrics []model.Metrics
	message, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	err = json.Unmarshal(message, &metrics)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal request body"})
		return
	}

	var gauges []storage.Valuer[float64]
	var counters []storage.Valuer[int64]
	for _, metric := range metrics {
		if metric.Type == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The metric type is empty"})
			return
		}

		if metric.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The metric id is empty"})
			return
		}

		switch metric.Type {
		case h.gaugeStorage.GetName():
			if metric.Value == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "The metric value is empty"})
				return
			}
			gauges = append(gauges, storage.Value[float64]{Key: metric.ID, Value: *metric.Value})
		case h.counterStorage.GetName():
			if metric.Delta == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "The metric delta is empty"})
				return
			}
			counters = append(counters, storage.Value[int64]{Key: metric.ID, Value: *metric.Delta})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metrics type"})
			return
		}
	}

	if len(gauges) > 0 {
		if err := retry.Retry(
			func() error {
				return h.gaugeStorage.Set(ctx, gauges...)
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set metric value"})
			return
		}

	}

	if len(counters) > 0 {
		if err := retry.Retry(
			func() error {
				return h.counterStorage.Set(ctx, counters...)
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set metric value"})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleValue(c *gin.Context) {
	ctx := c.Request.Context()

	var metricType string
	if metricType = c.Param("type"); metricType == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = c.Param("name"); metricName == "" {
		c.Status(http.StatusNotFound)
		return
	}

	var value string
	switch metricType {
	case h.gaugeStorage.GetName():
		if err := retry.Retry(
			func() (err error) {
				value, err = h.gaugeStorage.GetString(ctx, metricName)
				return
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			return
		}

		c.String(http.StatusOK, value)
		return

	case h.counterStorage.GetName():
		if err := retry.Retry(
			func() (err error) {
				value, err = h.counterStorage.GetString(ctx, metricName)
				return
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			return
		}

		c.String(http.StatusOK, value)
		return

	default:
		c.Status(http.StatusBadRequest)
		return
	}
}

func (h *Handler) handleValueJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var metric model.Metrics
	message, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	err = json.Unmarshal(message, &metric)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal request body"})
		return
	}

	if metric.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The metric type is empty"})
		return
	}

	if metric.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "The metric id is empty"})
		return
	}

	switch metric.Type {
	case h.gaugeStorage.GetName():
		var value float64
		if err := retry.Retry(
			func() (err error) {
				value, err = h.gaugeStorage.Get(ctx, metric.ID)
				return
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get metric value"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metric value"})
			return
		}
		metric.Value = &value

	case h.counterStorage.GetName():
		var value int64
		if err := retry.Retry(
			func() (err error) {
				value, err = h.counterStorage.Get(ctx, metric.ID)
				return
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get metric value"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metric value"})
			return
		}
		metric.Delta = &value

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric type"})
		return
	}

	c.JSON(http.StatusOK, metric)
}

func (h *Handler) handleHTML(c *gin.Context) {
	ctx := c.Request.Context()

	var gaugesStats, counterStats map[string]string
	if err := retry.Retry(
		func() (err error) {
			gaugesStats, err = h.gaugeStorage.GetAllString(ctx)
			return
		},
		func(err error) bool {
			return errors.Is(err, driver.ErrBadConn)
		},
		[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := retry.Retry(
		func() (err error) {
			counterStats, err = h.counterStorage.GetAllString(ctx)
			return
		},
		func(err error) bool {
			return errors.Is(err, driver.ErrBadConn)
		},
		[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	tmpl := `<html>
		<head>
		<title></title>
		</head>
		<body>
			<table border="1" cellpadding="1" cellspacing="1" style="width: 600px">
				<thead>
					<tr>
						<th scope="col" style="width: 300px">Gauge View</th>
						<th scope="col" style="width: 300px">Value</th>
					</tr>
				</thead>
				<tbody>
					%s
				</tbody>
			</table>
			<table border="1" cellpadding="1" cellspacing="1" style="width: 600px; margin-top: 30px;">
				<thead>
					<tr>
						<th scope="col" style="width: 300px">Counter View</th>
						<th scope="col" style="width: 300px">Value</th>
					</tr>
				</thead>
				<tbody>
					%s
				</tbody>
			</table>
		</body>
	</html>`

	gaugesView := h.mapToHTML(gaugesStats)
	counterView := h.mapToHTML(counterStats)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(tmpl, gaugesView, counterView)))
}

func (h *Handler) mapToHTML(m map[string]string) string {
	output := ""
	for k, v := range m {
		output += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", k, v)
	}
	return output
}

func (h *Handler) ping(c *gin.Context) {
	err := h.db.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to database"})
		return
	}
	c.String(http.StatusOK, "pong")
}
