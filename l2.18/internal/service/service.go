package service

import (
	"context"
	"fmt"
	"time"

	"calendar-wbf/internal/entity"
	"calendar-wbf/pkg/cache"
	"calendar-wbf/pkg/logger"
)

const (
	_defaultContextTimeout = 500 * time.Millisecond
)

type (
	EventRepo interface {
		Create(ctx context.Context, event *entity.Event) (*entity.Event, error)
		GetByID(ctx context.Context, id uint64) (*entity.Event, error)
		Update(ctx context.Context, event *entity.Event) (*entity.Event, error)
		Delete(ctx context.Context, id uint64) error
		GetByUserAndDate(ctx context.Context, userID uint64, date time.Time) ([]*entity.Event, error)
		GetByUserAndDateRange(ctx context.Context, userID uint64, startDate, endDate time.Time) ([]*entity.Event, error)
	}

	EventService struct {
		eventRepo EventRepo
		logger    logger.Logger
		cache     cache.Cache[uint64, *entity.Event]
		cacheTTL  time.Duration
	}
)

func NewEventService(
	eventRepo EventRepo,
	logger logger.Logger,
	cache cache.Cache[uint64, *entity.Event],
	cacheTTL time.Duration,
) *EventService {
	cache.SetOnEvicted(func(key uint64, value *entity.Event) {
		logger.Infow("cache eviction",
			"key", key,
			"event_id", value.ID,
		)
	})

	return &EventService{
		eventRepo: eventRepo,
		logger:    logger,
		cache:     cache,
		cacheTTL:  cacheTTL,
	}
}

func (s *EventService) CreateEvent(
	ctx context.Context,
	userID uint64,
	date time.Time,
	title, text string,
) (*entity.Event, error) {
	const op = "service.CreateEvent"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "create event started",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.String("date", date.String()),
		logger.String("title", title),
	)

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if duration > 200*time.Millisecond {
			log.LogAttrs(ctx, logger.WarnLevel, "slow service operation",
				logger.String("op", op),
				logger.Uint64("user_id", userID),
				logger.String("duration", duration.String()),
			)
		}
	}()

	event := &entity.Event{
		UserID: userID,
		Date:   date,
		Title:  title,
		Text:   text,
	}

	if err := s.validateEvent(event); err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "event validation failed",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("%s: validate event: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	createdEvent, err := s.eventRepo.Create(ctx, event)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "event creation failed",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("%s: create event: %w", op, err)
	}

	s.cache.Put(createdEvent.ID, createdEvent, s.cacheTTL)

	duration := time.Since(startTime)
	log.LogAttrs(ctx, logger.InfoLevel, "event created successfully",
		logger.String("op", op),
		logger.Uint64("event_id", createdEvent.ID),
		logger.String("duration", duration.String()),
	)

	return createdEvent, nil
}

func (s *EventService) UpdateEvent(
	ctx context.Context,
	id, userID uint64,
	date time.Time,
	title, text string,
) (*entity.Event, error) {
	const op = "service.UpdateEvent"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "update event started",
		logger.String("op", op),
		logger.Uint64("event_id", id),
		logger.Uint64("user_id", userID),
	)

	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "failed to get existing event",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("event_id", id),
		)
		return nil, fmt.Errorf("%s: get event: %w", op, err)
	}

	if existing.UserID != userID {
		log.LogAttrs(ctx, logger.WarnLevel, "unauthorized update attempt",
			logger.String("op", op),
			logger.Uint64("event_id", id),
			logger.Uint64("owner_id", existing.UserID),
			logger.Uint64("requested_by", userID),
		)
		return nil, fmt.Errorf("%s: %w", op, entity.ErrEventNotFound)
	}

	event := &entity.Event{
		ID:     id,
		UserID: userID,
		Date:   date,
		Title:  title,
		Text:   text,
	}

	if validateErr := s.validateEvent(event); validateErr != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "event validation failed",
			logger.String("op", op),
			logger.Any("error", validateErr),
			logger.Uint64("event_id", id),
		)
		return nil, fmt.Errorf("%s: validate event: %w", op, validateErr)
	}

	updatedEvent, err := s.eventRepo.Update(ctx, event)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "event update failed",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("event_id", id),
		)
		return nil, fmt.Errorf("%s: update event: %w", op, err)
	}

	s.cache.Put(updatedEvent.ID, updatedEvent, s.cacheTTL)

	duration := time.Since(startTime)
	log.LogAttrs(ctx, logger.InfoLevel, "event updated successfully",
		logger.String("op", op),
		logger.Uint64("event_id", id),
		logger.String("duration", duration.String()),
	)

	return updatedEvent, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, id, userID uint64) error {
	const op = "service.DeleteEvent"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "delete event started",
		logger.String("op", op),
		logger.Uint64("event_id", id),
		logger.Uint64("user_id", userID),
	)

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "failed to get existing event",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("event_id", id),
		)
		return fmt.Errorf("%s: get event: %w", op, err)
	}

	if existing.UserID != userID {
		log.LogAttrs(ctx, logger.WarnLevel, "unauthorized delete attempt",
			logger.String("op", op),
			logger.Uint64("event_id", id),
			logger.Uint64("owner_id", existing.UserID),
			logger.Uint64("requested_by", userID),
		)
		return fmt.Errorf("%s: %w", op, entity.ErrEventNotFound)
	}

	if repoErr := s.eventRepo.Delete(ctx, id); repoErr != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "event deletion failed",
			logger.String("op", op),
			logger.Any("error", repoErr),
			logger.Uint64("event_id", id),
		)
		return fmt.Errorf("%s: delete event: %w", op, repoErr)
	}

	log.LogAttrs(ctx, logger.InfoLevel, "event deleted successfully",
		logger.String("op", op),
		logger.Uint64("event_id", id),
	)

	return nil
}

func (s *EventService) GetEventsForDay(ctx context.Context, userID uint64, date time.Time) ([]*entity.Event, error) {
	const op = "service.GetEventsForDay"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "get events for day requested",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.String("date", date.String()),
	)

	if err := s.validateUserID(userID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	events, err := s.eventRepo.GetByUserAndDate(ctx, userID, date)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "failed to get events",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.LogAttrs(ctx, logger.InfoLevel, "events retrieved successfully",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.Int("count", len(events)),
	)

	return events, nil
}

func (s *EventService) GetEventsForWeek(
	ctx context.Context,
	userID uint64,
	startDate time.Time,
) ([]*entity.Event, error) {
	const op = "service.GetEventsForWeek"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "get events for week requested",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.String("start_date", startDate.String()),
	)

	if err := s.validateUserID(userID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// nolint: mnd
	endDate := startDate.AddDate(0, 0, 6)

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	events, err := s.eventRepo.GetByUserAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "failed to get events",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.LogAttrs(ctx, logger.InfoLevel, "events retrieved successfully",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.Int("count", len(events)),
	)

	return events, nil
}

func (s *EventService) GetEventsForMonth(ctx context.Context, userID uint64, year, month int) ([]*entity.Event, error) {
	const op = "service.GetEventsForMonth"
	log := s.logger.Ctx(ctx)

	log.LogAttrs(ctx, logger.InfoLevel, "get events for month requested",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.Int("year", year),
		logger.Int("month", month),
	)

	if err := s.validateUserID(userID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if year < 1900 || year > 2100 {
		return nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidDate)
	}

	if month < 1 || month > 12 {
		return nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidDate)
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	ctx, cancel := context.WithTimeout(ctx, _defaultContextTimeout)
	defer cancel()

	events, err := s.eventRepo.GetByUserAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		log.LogAttrs(ctx, logger.ErrorLevel, "failed to get events",
			logger.String("op", op),
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.LogAttrs(ctx, logger.InfoLevel, "events retrieved successfully",
		logger.String("op", op),
		logger.Uint64("user_id", userID),
		logger.Int("count", len(events)),
	)

	return events, nil
}

func (s *EventService) validateEvent(event *entity.Event) error {
	if event.ID == 0 {
		return entity.ErrInvalidDate
	}
	if event.Date.IsZero() {
		return entity.ErrInvalidDate
	}
	if event.Title == "" {
		return entity.ErrInvalidData
	}
	if event.Text == "" {
		return entity.ErrInvalidData
	}
	return nil
}

func (s *EventService) validateUserID(userID uint64) error {
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	return nil
}
