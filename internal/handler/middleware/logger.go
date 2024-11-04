package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger(c *gin.Context) {
	start := time.Now()

	var buf bytes.Buffer
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	tee := io.TeeReader(c.Request.Body, &buf)
	body, _ := io.ReadAll(tee)
	c.Request.Body = io.NopCloser(&buf)

	c.Next()

	fields := []logger.Field{
		logger.Any("start", start),
		logger.Any("latency", time.Since(start).String()),
		logger.Any("method", c.Request.Method),
		logger.Any("status", c.Writer.Status()),
		logger.Any("client_ip", c.ClientIP()),
		logger.Any("request", string(body)),
		logger.Any("response", blw.body.String()),
	}

	logger.Log.Info(c.Request.RequestURI, fields...)
}
