package httpt

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"calendar-wbf/internal/entity"
	"calendar-wbf/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	_defaultContextTimeout = 500 * time.Millisecond
)

type EventService interface {
	CreateEvent(ctx context.Context, userID uint64, date time.Time, title, text string) (*entity.Event, error)
	UpdateEvent(ctx context.Context, id, userID uint64, date time.Time, title, text string) (*entity.Event, error)
	DeleteEvent(ctx context.Context, id, userID uint64) error
	GetEventsForDay(ctx context.Context, userID uint64, date time.Time) ([]*entity.Event, error)
	GetEventsForWeek(ctx context.Context, userID uint64, startDate time.Time) ([]*entity.Event, error)
	GetEventsForMonth(ctx context.Context, userID uint64, year, month int) ([]*entity.Event, error)
}

// @Summary Создать событие
// @Description Создает новое событие в календаре
// @Tags Events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Данные для создания события"
// @Success 200 {object} entity.Event "Успешный ответ с данными события"
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /create_event [post]
func (h *CalendarHandler) createEventHandler(c *gin.Context) {
	const op = "transport.createEventHandler"
	log := h.log.Ctx(c.Request.Context())

	var req CreateEventRequest
	if svcErr := c.ShouldBindJSON(&req); svcErr != nil {
		h.handleBindError(c, svcErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	event, err := h.svc.CreateEvent(ctx, req.UserID, req.Date, req.Title, req.Text)
	if err != nil {
		h.handleServiceError(c, err, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event created successfully",
		logger.Uint64("event_id", event.ID),
	)

	c.JSON(http.StatusOK, gin.H{"result": "Event create successfully"})
}

// @Summary Обновить событие
// @Description Обновляет существующее событие
// @Tags Events
// @Accept json
// @Produce json
// @Param id path int true "ID события"
// @Param request body UpdateEventRequest true "Данные для обновления события"
// @Success 200 {object} entity.Event "Успешный ответ с данными события"
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 404 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /update_event/{id} [post]
func (h *CalendarHandler) updateEventHandler(c *gin.Context) {
	const op = "transport.updateEventHandler"
	log := h.log.Ctx(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.handleEventIDError(c, op, id)
		return
	}

	var req UpdateEventRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		h.handleBindError(c, bindErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	event, err := h.svc.UpdateEvent(ctx, id, req.UserID, req.Date, req.Title, req.Text)
	if err != nil {
		h.handleServiceError(c, err, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event update successfully",
		logger.Uint64("event_id", event.ID),
	)

	c.JSON(http.StatusOK, gin.H{"result": "Event update successfully"})
}

// @Summary Удалить событие
// @Description Удаляет событие из календаря
// @Tags Events
// @Accept json
// @Produce json
// @Param id path int true "ID события"
// @Param request body DeleteEventRequest true "Данные для удаления события"
// @Success 200 {object} gin.H "Успешное удаление"
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 404 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /delete_event/{id} [post]
func (h *CalendarHandler) deleteEventHandler(c *gin.Context) {
	const op = "transport.deleteEventHandler"
	log := h.log.Ctx(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.handleEventIDError(c, op, id)
		return
	}

	var req DeleteEventRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		h.handleBindError(c, bindErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	if svcErr := h.svc.DeleteEvent(ctx, id, req.UserID); svcErr != nil {
		h.handleServiceError(c, svcErr, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event deleted successfully",
		logger.Uint64("event_id", id),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// @Summary Получить события на день
// @Description Получает список событий календаря на указанный день для заданного пользователя
// @Tags Events
// @Accept json
// @Produce json
// @Param request body httpt.GetEventForDayRequest true "Данные запроса: UserID и Date"
// @Success 200 {object} httpt.EventsResponse
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /events_for_day [get]
func (h *CalendarHandler) getEventsForDayHandler(c *gin.Context) {
	const op = "transport.getEventsForDayHandler"
	log := h.log.Ctx(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.handleEventIDError(c, op, id)
		return
	}

	var req GetEventForDayRequest
	if svcErr := c.ShouldBindJSON(&req); svcErr != nil {
		h.handleBindError(c, svcErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	event, err := h.svc.GetEventsForDay(ctx, req.UserID, req.Date)
	if err != nil {
		h.handleServiceError(c, err, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event for day retreived successfully",
		logger.Slice("events", event),
	)

	c.JSON(http.StatusOK, gin.H{"result": "Event deleted successfully"})
}

// @Summary Получить события на неделю
// @Description Получает список событий календаря на неделю, начиная с указанной даты
// @Tags Events
// @Accept json
// @Produce json
// @Param request body GetEventForWeekRequest true "Данные запроса: UserID и StartDate"
// @Success 200 {object} httpt.EventsResponse
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /events_for_week [get]
func (h *CalendarHandler) getEventsForWeekHandler(c *gin.Context) {
	const op = "transport.getEventsForWeekHandler"
	log := h.log.Ctx(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.handleEventIDError(c, op, id)
		return
	}

	var req GetEventForWeekRequest
	if svcErr := c.ShouldBindJSON(&req); svcErr != nil {
		h.handleBindError(c, svcErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	event, err := h.svc.GetEventsForWeek(ctx, req.UserID, req.StartDate)
	if err != nil {
		h.handleServiceError(c, err, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event for week retreived successfully",
		logger.Slice("events", event),
	)

	c.JSON(http.StatusOK, event)
}

// @Summary Получить события на месяц
// @Description Получает список событий календаря на указанный месяц и год
// @Tags Events
// @Accept json
// @Produce json
// @Param request body GetEventForMonthRequest true "Данные запроса: UserID, Year и Month"
// @Success 200 {object} httpt.EventsResponse
// @Failure 400 {object} httpt.ErrorResponse
// @Failure 500 {object} httpt.ErrorResponse
// @Router /events_for_month [get]
func (h *CalendarHandler) getEventsForMonthsHandler(c *gin.Context) {
	const op = "transport.getEventsForMonthsHandler"
	log := h.log.Ctx(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.handleEventIDError(c, op, id)
		return
	}

	var req GetEventForMonthRequest
	if svcErr := c.ShouldBindJSON(&req); svcErr != nil {
		h.handleBindError(c, svcErr, op)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), _defaultContextTimeout)
	defer cancel()

	event, err := h.svc.GetEventsForMonth(ctx, req.UserID, req.Year, req.Month)
	if err != nil {
		h.handleServiceError(c, err, op)
		return
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event for month retreived successfully",
		logger.Slice("events", event),
	)

	c.JSON(http.StatusOK, event)
}
