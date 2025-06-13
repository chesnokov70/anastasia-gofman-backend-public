package middleware

import (
	"bytes"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func DetailedLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		log.Printf("[REQUEST] %s %s", c.Request.Method, path)
		if raw != "" {
			log.Printf("[QUERY] %s", raw)
		}

		log.Printf("[HEADERS] Content-Type: %s", c.GetHeader("Content-Type"))
		log.Printf("[HEADERS] Content-Length: %s", c.GetHeader("Content-Length"))

		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := c.Request.ParseMultipartForm(32 << 20); err == nil {
				if c.Request.MultipartForm != nil {
					log.Printf("[MULTIPART] Form values: %v", c.Request.MultipartForm.Value)
					for fieldName, files := range c.Request.MultipartForm.File {
						log.Printf("[MULTIPART] Field '%s' has %d file(s)", fieldName, len(files))
						for i, file := range files {
							log.Printf("[MULTIPART] File %d: %s (size: %d bytes)", i+1, file.Filename, file.Size)
						}
					}
				}
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Printf("[RESPONSE] %s %s - Status: %d, Latency: %v",
			c.Request.Method,
			path,
			statusCode,
			latency)

		if len(c.Errors) > 0 {
			log.Printf("[ERRORS] %v", c.Errors.String())
		}
	}
}
