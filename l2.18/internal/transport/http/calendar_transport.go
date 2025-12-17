package httpt

import (
	"calendar-wbf/pkg/logger"

	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	svc    EventService
	log    logger.Logger
	router *gin.Engine
}

func NewCalendarHandler(
	svc EventService,
	log logger.Logger,
) *CalendarHandler {
	h := &CalendarHandler{
		svc: svc,
		log: log,
	}

	router := gin.New()

	router.Use(h.requestIDMiddleware())
	router.Use(h.loggingMiddleware())
	router.Use(gin.Recovery())

	h.router = router

	h.router.LoadHTMLGlob("web/*.html")
	h.router.Static("/static", "./web")

	h.setupRoutes()

	return h
}

func (h *CalendarHandler) Engine() *gin.Engine {
	return h.router
}
