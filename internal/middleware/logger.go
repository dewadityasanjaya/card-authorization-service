package middleware

import (
	"time"

	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger logs every incoming HTTP request
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request completes
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}
