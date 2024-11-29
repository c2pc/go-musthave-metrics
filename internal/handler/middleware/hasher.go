package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const hashHeader = "HashSHA256"

type Hasher interface {
	Check([]byte, []byte) (bool, error)
	Hash([]byte) (string, error)
}

type hasherResponseWriter struct {
	gin.ResponseWriter
	hasher Hasher
}

func (grw *hasherResponseWriter) Write(data []byte) (int, error) {
	hash, err := grw.hasher.Hash(data)
	if err != nil {
		return 0, err
	}

	grw.Header().Set(hashHeader, hash)

	return grw.ResponseWriter.Write(data)
}

func HashMiddleware(hasher Hasher) func(*gin.Context) {
	var needHash bool
	if hasher != nil {
		needHash = true
	}

	return func(c *gin.Context) {
		requestHash := c.Request.Header.Get(hashHeader)
		if requestHash != "" && needHash {
			var buf bytes.Buffer
			blw := &hasherResponseWriter{ResponseWriter: c.Writer, hasher: hasher}
			c.Writer = blw

			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)

			ok, err := hasher.Check(body, []byte(requestHash))
			if err != nil || !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to check hash"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
