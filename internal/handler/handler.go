package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/c2pc/go-musthave-metrics/internal/handler/middleware"
	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/gin-gonic/gin"
)

type Storager[T int64 | float64] interface {
	GetName() string
	Get(key string) (T, error)
	GetString(key string) (string, error)
	GetAll() (map[string]T, error)
	GetAllString() (map[string]string, error)
	Set(key string, value T) error
	SetString(key string, value string) error
}

type DBDriver interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	http.Handler
	gaugeStorage   Storager[float64]
	counterStorage Storager[int64]
	db             DBDriver
}

func NewHandler(gaugeStorage Storager[float64], counterStorage Storager[int64], db DBDriver) (http.Handler, error) {
	if counterStorage == nil {
		return nil, fmt.Errorf("counterStorage is empty")
	}

	if gaugeStorage == nil {
		return nil, fmt.Errorf("gaugeStorage is empty")
	}

	if db == nil {
		return nil, fmt.Errorf("database is empty")
	}

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

	return h, nil
}

func (h *Handler) Init(engine *gin.Engine) {
	api := engine.Group("", middleware.GzipDecompressor, middleware.GzipCompressor, middleware.Logger)
	{
		api.GET("/", h.handleHTML)
		api.GET("/ping", h.ping)
		api.POST("/update/", h.handleUpdateJSON)
		api.POST("/update/:type/:name/:value", h.handleUpdate)
		api.GET("/value/:type/:name", h.handleValue)
		api.POST("/value/", h.handleValueJSON)
	}
}

func (h *Handler) handleUpdate(c *gin.Context) {
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
		if err := h.gaugeStorage.SetString(metricName, metricValue); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

	case h.counterStorage.GetName():
		if err := h.counterStorage.SetString(metricName, metricValue); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

	default:
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleUpdateJSON(c *gin.Context) {
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

		if err := h.gaugeStorage.Set(metric.ID, *metric.Value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to set metric value"})
			return
		}

		newValue, err := h.gaugeStorage.Get(metric.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get metric value"})
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

		if err := h.counterStorage.Set(metric.ID, *metric.Delta); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to set metric value"})
			return
		}

		newValue, err := h.counterStorage.Get(metric.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get metric value"})
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

func (h *Handler) handleValue(c *gin.Context) {
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

	switch metricType {
	case h.gaugeStorage.GetName():
		value, err := h.gaugeStorage.GetString(metricName)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, value)
		return

	case h.counterStorage.GetName():
		value, err := h.counterStorage.GetString(metricName)
		if err != nil {
			c.Status(http.StatusNotFound)
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
		value, err := h.gaugeStorage.Get(metric.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get metric value"})
			return
		}
		metric.Value = &value

	case h.counterStorage.GetName():
		value, err := h.counterStorage.Get(metric.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get metric value"})
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
	gaugesStats, err := h.gaugeStorage.GetAllString()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	counterStats, err := h.counterStorage.GetAllString()
	if err != nil {
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
	ctx := c.Request.Context()
	if h.db == nil {
		c.Status(http.StatusServiceUnavailable)
	}

	err := h.db.Ping(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to database"})
		return
	}
	c.String(http.StatusOK, "pong")
}
