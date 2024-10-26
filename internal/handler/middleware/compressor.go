package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (grw *gzipResponseWriter) Write(data []byte) (int, error) {
	return grw.Writer.Write(data)
}

func GzipDecompressor(c *gin.Context) {
	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		gzipData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
			c.Abort()
			return
		}

		reader, err := gzip.NewReader(bytes.NewReader(gzipData))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create gzip reader"})
			c.Abort()
			return
		}
		defer reader.Close()

		decompressedData, err := io.ReadAll(reader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decompress gzip data"})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(decompressedData))
		c.Request.ContentLength = int64(len(decompressedData))
		c.Request.Header.Set("Content-Encoding", "")
		c.Request.Header.Set("Content-Type", http.DetectContentType(decompressedData))
	}

	c.Next()
}

func GzipCompressor(c *gin.Context) {
	acceptEncoding := c.Request.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "gzip") {
		gz, err := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		if err != nil {
			return
		}

		defer gz.Reset(io.Discard)
		gz.Reset(c.Writer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &gzipResponseWriter{c.Writer, gz}
		defer func() {
			gz.Close()
			c.Header("Content-Length", fmt.Sprint(c.Writer.Size()))
		}()
		c.Next()
		return
	}
	c.Next()
}
