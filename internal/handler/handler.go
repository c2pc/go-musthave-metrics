package handler

import (
	"fmt"
	"net/http"

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

type Handler struct {
	http.Handler
	gaugeStorage   Storager[float64]
	counterStorage Storager[int64]
}

func NewHandler(gaugeStorage Storager[float64], counterStorage Storager[int64]) (http.Handler, error) {
	if counterStorage == nil {
		return nil, fmt.Errorf("counterStorage is nil")
	}

	if gaugeStorage == nil {
		return nil, fmt.Errorf("gaugeStorage is nil")
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
	}

	h.Init(handlers)

	return h, nil
}

func (h *Handler) Init(engine *gin.Engine) {
	api := engine.Group("")
	{
		api.GET("/", h.handleHTML)
		api.POST("/update/:type/:name/:value", h.handleUpdate)
		api.GET("/value/:type/:name", h.handleValue)
	}
}

func (h *Handler) handleUpdate(c *gin.Context) {
	var metricType string
	if metricType = c.Param("type"); metricType == "" {
		fmt.Println("metric type is empty")
		c.Status(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = c.Param("name"); metricName == "" {
		fmt.Println("metric name is empty")
		c.Status(http.StatusNotFound)
		return
	}

	var metricValue string
	if metricValue = c.Param("value"); metricValue == "" {
		fmt.Println("metric value is empty")
		c.Status(http.StatusBadRequest)
		return
	}

	fmt.Println("Request ", metricType, metricName, metricValue)

	switch metricType {
	case h.gaugeStorage.GetName():
		if err := h.gaugeStorage.SetString(metricName, metricValue); err != nil {
			fmt.Println("could not store metric value", err)
			c.Status(http.StatusBadRequest)
			return
		}

	case h.counterStorage.GetName():
		if err := h.counterStorage.SetString(metricName, metricValue); err != nil {
			fmt.Println("could not store metric value", err)
			c.Status(http.StatusBadRequest)
			return
		}

	default:
		fmt.Println("unknown metric type")
		c.Status(http.StatusBadRequest)
		return
	}

	fmt.Println("Success ", metricType, metricName, metricValue)

	c.Status(http.StatusOK)
}

func (h *Handler) handleValue(c *gin.Context) {
	var metricType string
	if metricType = c.Param("type"); metricType == "" {
		fmt.Println("metric type is empty")
		c.Status(http.StatusBadRequest)
		return
	}

	var metricName string
	if metricName = c.Param("name"); metricName == "" {
		fmt.Println("metric name is empty")
		c.Status(http.StatusNotFound)
		return
	}

	fmt.Println("Request ", metricType, metricName)

	switch metricType {
	case h.gaugeStorage.GetName():
		value, err := h.gaugeStorage.GetString(metricName)
		if err != nil {
			fmt.Println("not found store metric value", err)
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, value)
		return
	case h.counterStorage.GetName():
		value, err := h.counterStorage.GetString(metricName)
		if err != nil {
			fmt.Println("not found metric value", err)
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, value)
		return
	default:
		fmt.Println("unknown metric type")
		c.Status(http.StatusBadRequest)
		return
	}
}

func (h *Handler) handleHTML(c *gin.Context) {
	gaugesStats, err := h.gaugeStorage.GetAllString()
	if err != nil {
		fmt.Println("could not get gauge stats", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	counterStats, err := h.counterStorage.GetAllString()
	if err != nil {
		fmt.Println("could not get counter stats", err)
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
