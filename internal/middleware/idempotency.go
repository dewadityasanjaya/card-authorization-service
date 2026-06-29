package middleware

import (
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IdempotencyKey extracts and stores the idempotency key from the header
func IdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key != "" {
			logger.Info("Idempotency key received",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
			)
			c.Set("idempotency_key", key)
		}
		c.Next()
	}
}
