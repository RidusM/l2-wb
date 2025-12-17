package repository

import (
	"context"
	"sync"
	"time"

	"calendar-wbf/internal/entity"
)

type EventRepository struct {
	mu        sync.RWMutex
	events    map[uint64]*entity.Event
	nextID    uint64
	userIndex map[uint64]map[string][]uint64
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events:    make(map[uint64]*entity.Event),
		nextID:    1,
		userIndex: make(map[uint64]map[string][]uint64),
	}
}

func (r *EventRepository) Create(_ context.Context, event *entity.Event) (*entity.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	event.ID = r.nextID
	event.CreatedAt = now
	event.UpdatedAt = now
	r.nextID++

	r.events[event.ID] = event
	r.addToIndex(event)

	return event, nil
}

func (r *EventRepository) GetByID(_ context.Context, id uint64) (*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, exists := r.events[id]
	if !exists {
		return nil, entity.ErrEventNotFound
	}

	return event, nil
}

func (r *EventRepository) Update(_ context.Context, event *entity.Event) (*entity.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.events[event.ID]
	if !exists {
		return nil, entity.ErrEventNotFound
	}

	r.removeFromIndex(existing)

	event.CreatedAt = existing.CreatedAt
	event.UpdatedAt = time.Now()
	r.events[event.ID] = event

	r.addToIndex(event)

	return event, nil
}

func (r *EventRepository) Delete(_ context.Context, id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, exists := r.events[id]
	if !exists {
		return entity.ErrEventNotFound
	}

	r.removeFromIndex(event)
	delete(r.events, id)

	return nil
}

func (r *EventRepository) GetByUserAndDate(
	_ context.Context,
	userID uint64,
	date time.Time,
) ([]*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dateKey := date.String()
	return r.getEventsByDateKey(userID, dateKey), nil
}

func (r *EventRepository) GetByUserAndDateRange(
	_ context.Context,
	userID uint64,
	startDate, endDate time.Time,
) ([]*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.Event
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.String()
		events := r.getEventsByDateKey(userID, dateKey)
		result = append(result, events...)
	}

	return result, nil
}

func (r *EventRepository) addToIndex(event *entity.Event) {
	if r.userIndex[event.UserID] == nil {
		r.userIndex[event.UserID] = make(map[string][]uint64)
	}

	dateKey := event.Date.String()
	r.userIndex[event.UserID][dateKey] = append(r.userIndex[event.UserID][dateKey], event.ID)
}

func (r *EventRepository) removeFromIndex(event *entity.Event) {
	dateKey := event.Date.String()
	if userDates, exists := r.userIndex[event.UserID]; exists {
		if ids, idsExists := userDates[dateKey]; idsExists {
			for i, id := range ids {
				if id == event.ID {
					r.userIndex[event.UserID][dateKey] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}
}

func (r *EventRepository) getEventsByDateKey(userID uint64, dateKey string) []*entity.Event {
	var result []*entity.Event

	if userDates, exists := r.userIndex[userID]; exists {
		if ids, idsExists := userDates[dateKey]; idsExists {
			for _, id := range ids {
				if event, eventExists := r.events[id]; eventExists {
					result = append(result, event)
				}
			}
		}
	}

	return result
}
