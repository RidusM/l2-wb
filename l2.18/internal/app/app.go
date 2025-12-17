package app

import (
	"context"
	"fmt"

	"calendar-wbf/internal/config"
	"calendar-wbf/internal/entity"
	"calendar-wbf/internal/repository"
	"calendar-wbf/internal/service"
	httpt "calendar-wbf/internal/transport/http"
	"calendar-wbf/pkg/cache"
	"calendar-wbf/pkg/logger"

	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg *config.Config, log logger.Logger) error {
	eg, ctx := errgroup.WithContext(ctx)

	calendarCache, err := initCache(&cfg.Cache, log)
	if err != nil {
		return err
	}
	defer stopCache(calendarCache)

	calendarService := initEventService(
		cfg,
		calendarCache,
		log,
	)

	if serverErr := initHTTPServer(ctx, eg, &cfg.HTTP, calendarService, log); serverErr != nil {
		return serverErr
	}

	return waitForShutdown(eg)
}

func initCache(
	cfg *config.Cache,
	log logger.Logger,
) (cache.Cache[uint64, *entity.Event], error) {
	calendarCache, err := cache.NewLRUCache[uint64, *entity.Event](
		cfg.Capacity,
		log.With("component", "cache"),
	)
	if err != nil {
		return nil, fmt.Errorf("app.initCache: %w", err)
	}
	calendarCache.StartCleanup(cfg.CleanupInterval)
	return calendarCache, nil
}

func stopCache(calendarCache cache.Cache[uint64, *entity.Event]) {
	if calendarCache != nil {
		calendarCache.StopCleanup()
	}
}

func initEventService(
	cfg *config.Config,
	calendarCache cache.Cache[uint64, *entity.Event],
	log logger.Logger,
) *service.EventService {
	calendarRepo := repository.NewEventRepository()

	calendarService := service.NewEventService(
		calendarRepo,
		log.With("component", "calendar service"),
		calendarCache,
		cfg.Cache.TTL,
	)

	return calendarService
}

func initHTTPServer(
	ctx context.Context,
	eg *errgroup.Group,
	cfg *config.HTTP,
	calendarService *service.EventService,
	log logger.Logger,
) error {
	httpServer, err := httpt.NewHTTPServer(
		httpt.NewCalendarHandler(calendarService, log),
		cfg,
		log.With("component", "http server"),
	)
	if err != nil {
		return fmt.Errorf("app.initHTTPServer: %w", err)
	}

	eg.Go(func() error {
		return httpServer.Start(ctx)
	})
	return nil
}

func waitForShutdown(eg *errgroup.Group) error {
	if err := eg.Wait(); err != nil && !isShutdownSignal(err) {
		return fmt.Errorf("app.waitForShutdown: application failed: %w", err)
	}
	return nil
}

func isShutdownSignal(err error) bool {
	return err != nil && err.Error() == "shutdown signal"
}
