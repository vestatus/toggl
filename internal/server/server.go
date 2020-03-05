package server

import (
	"context"
	"time"

	"github.com/vestatus/toggl/internal/logger"
	"github.com/vestatus/toggl/internal/service"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	Config
	service *service.Service
}

func New(config Config, svc *service.Service) *Server {
	return &Server{
		Config:  config,
		service: svc,
	}
}

func (s *Server) loadTakers(ctx context.Context) error {
	// I must note that it would probably be better to expose a handler and
	// have another service control the timing (since we might want to scale horizontally)
	ticker := time.NewTicker(s.UpdateInterval)
	defer ticker.Stop()

	log := logger.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := s.service.LoadTakers(ctx)
			if service.IsFatal(err) {
				return err
			}
			if err != nil {
				log.WithError(err).Error("failed to load takers")
			}
		}
	}
}

func (s *Server) sendThanks(ctx context.Context) error {
	log := logger.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			timeoutCtx, _ := context.WithTimeout(ctx, s.SendThanksTimeout)

			ok, err := s.service.SendNextThanks(timeoutCtx)
			if service.IsFatal(err) {
				return err
			}
			if err != nil {
				log.WithError(err).Error("failed to send thanks")
				continue
			}

			if ok {
				continue
			}

			logger.FromContext(ctx).Debugf("sender will sleep for %s", s.PollInterval.String())

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.PollInterval):
			}
		}
	}
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.loadTakers(ctx)
	})
	eg.Go(func() error {
		return s.sendThanks(ctx)
	})

	return eg.Wait()
}
