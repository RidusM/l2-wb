package httpt

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Calendar Service API
// @version         1.0
// @description     API для управления календарём
// @termsOfService  http://swagger.io/terms/
// @contact.name    RidusM
// @contact.email   stormkillpeople@gmail.com
// @license.name    MIT-0
// @license.url     https://github.com/aws/mit-0
// @host            localhost:8080
// @BasePath        /
func (h *CalendarHandler) setupRoutes() {
	h.router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	h.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	h.router.POST("/create_event", h.createEventHandler)
	h.router.POST("/update_event", h.updateEventHandler)
	h.router.POST("/delete_event", h.deleteEventHandler)
	h.router.POST("/events_for_day", h.getEventsForDayHandler)
	h.router.POST("/events_for_week", h.getEventsForWeekHandler)
	h.router.POST("/events_for_month", h.getEventsForMonthsHandler)

	h.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
