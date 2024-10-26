package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	fields := []zapcore.Field{
		zap.Time("start", start),
		zap.String("latency", time.Since(start).String()),
		zap.String("method", c.Request.Method),
		zap.Int("status", c.Writer.Status()),
		zap.String("client_ip", c.ClientIP()),
		zap.String("request", string(body)),
		zap.String("response", blw.body.String()),
	}

	logger.Log.Info(c.Request.RequestURI, fields...)
}
