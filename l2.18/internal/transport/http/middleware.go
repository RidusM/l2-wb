package httpt

import (
	"time"

	"calendar-wbf/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h *CalendarHandler) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := h.log.GenerateRequestID()
		ctx := h.log.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

func (h *CalendarHandler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		h.log.LogAttrs(c.Request.Context(), logger.InfoLevel, "HTTP request",
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.String("duration", latency.String()),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
		)
	}
}
