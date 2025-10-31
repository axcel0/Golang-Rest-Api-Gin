package middleware

import (
	"time"

	"Go-Lang-project-01/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Logger middleware for structured logging with slog
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Process request
		c.Next()

		// Calculate latency and log with structured fields
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Determine log level based on status code
		logFunc := logger.Info
		if statusCode >= 500 {
			logFunc = logger.Error
		} else if statusCode >= 400 {
			logFunc = logger.Warn
		}

		logFunc("HTTP Request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"client_ip", clientIP,
			"user_agent", c.Request.UserAgent(),
		)
	}
}
