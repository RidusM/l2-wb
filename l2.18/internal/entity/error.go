package entity

import (
	"errors"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrInvalidData      = errors.New("invalid data")
	ErrInvalidDate      = errors.New("invalid date")
	ErrInvalidUserID    = errors.New("invalid user_id")
	ErrDuplicateEvent   = errors.New("event already exists")
	ErrConfigPathNotSet = errors.New("CONFIG_PATH not set and -config flag not provided")
)
