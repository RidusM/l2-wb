// nolint: revive,staticcheck
package httpt

import (
	"time"

	"calendar-wbf/internal/entity"
)

// swagger: model ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}

// swagger: model Event
type Entity entity.Event

// swagger: model EventsResponse
type EventsResponse struct {
	Result []entity.Event `json:"result"`
}

// swagger: model CreateEventRequest
type CreateEventRequest struct {
	UserID uint64    `json:"user_id" binding:"required,gt=0"`
	Date   time.Time `json:"date"    binding:"required"`
	Title  string    `json:"title"   binding:"required"`
	Text   string    `json:"text"`
}

// swagger: model UpdateEventRequest
type UpdateEventRequest struct {
	UserID uint64    `json:"user_id" binding:"required,gt=0"`
	Date   time.Time `json:"date"    binding:"required"`
	Title  string    `json:"title"   binding:"required"`
	Text   string    `json:"text"`
}

// swagger: model DeleteEventRequest
type DeleteEventRequest struct {
	UserID uint64 `json:"user_id" binding:"required,gt=0"`
}

// swagger: model GetEventForDayRequest
type GetEventForDayRequest struct {
	UserID uint64    `json:"user_id" form:"user_id" binding:"required,gt=0"`
	Date   time.Time `json:"date"    form:"date"    binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

// swagger: model GetEventForWeekRequest
type GetEventForWeekRequest struct {
	UserID    uint64    `json:"user_id"    form:"user_id"    binding:"required,gt=0"`
	StartDate time.Time `json:"start_date" form:"start_date" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

// swagger: model GetEventForMonthRequest
type GetEventForMonthRequest struct {
	UserID uint64 `json:"user_id" form:"user_id" binding:"required,gt=0"`
	Year   int    `json:"year"    form:"year"    binding:"required"`
	Month  int    `json:"month"   form:"month"   binding:"required"`
}