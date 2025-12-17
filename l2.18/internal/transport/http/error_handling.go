package httpt

import (
	"context"
	"errors"
	"net/http"

	"calendar-wbf/internal/entity"
	"calendar-wbf/pkg/logger"

	"github.com/gin-gonic/gin"
)

func (h *CalendarHandler) handleServiceError(c *gin.Context, err error, op string) {
	log := h.log.Ctx(c.Request.Context())

	log.LogAttrs(c.Request.Context(), logger.ErrorLevel, op+" failed",
		logger.Any("error", err),
		logger.String("remote_addr", c.ClientIP()),
		logger.String("user_agent", c.Request.UserAgent()),
	)

	switch {
	case errors.Is(err, entity.ErrInvalidData):
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Invalid event data. Check title or text."},
		)
	case errors.Is(err, entity.ErrInvalidUserID):
		log.LogAttrs(c.Request.Context(), logger.WarnLevel, "event not found",
			logger.String("event_id", c.Param("event_id")),
			logger.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, context.DeadlineExceeded):
		log.LogAttrs(c.Request.Context(), logger.WarnLevel, "request timeout",
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
	default:
		log.LogAttrs(c.Request.Context(), logger.ErrorLevel, "internal server error",
			logger.Any("error", err),
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal service error"})
	}
}

func (h *CalendarHandler) handleEventIDError(c *gin.Context, op string, eventID uint64) {
	log := h.log.Ctx(c.Request.Context())

	log.LogAttrs(c.Request.Context(), logger.WarnLevel, "invalid event ID",
		logger.String("op", op),
		logger.Uint64("event_id", eventID),
		logger.String("client_ip", c.ClientIP()),
	)
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
}

func (h *CalendarHandler) handleBindError(c *gin.Context, err error, op string) {
	log := h.log.Ctx(c.Request.Context())

	log.LogAttrs(c.Request.Context(), logger.WarnLevel, "failed to bind request",
		logger.String("op", op),
		logger.Any("error", err),
		logger.String("client_ip", c.ClientIP()),
	)
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
}
