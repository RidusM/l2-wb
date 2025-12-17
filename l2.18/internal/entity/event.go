package entity

import "time"

type Event struct {
	ID        uint64    `json:"id"         validate:"required,gte=1"`
	UserID    uint64    `json:"user_id"    validate:"required,gte=1"`
	Date      time.Time `json:"date"       validate:"required"`
	Title     string    `json:"title"      validate:"required,max=50"`
	Text      string    `json:"text"       validate:"required,max=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}
