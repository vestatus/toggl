package server

import (
	"context"
	"log"
	"time"
	"toggl/internal/service"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	UpdateInterval time.Duration
	PollInterval   time.Duration
	Service        *service.Service
}

func (s *Server) loadTakers(ctx context.Context) error {
	// I must note that it would probably be better to expose a handler and
	// have another service control the timing (since we might want to scale horizontally)
	ticker := time.NewTicker(s.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := s.Service.LoadTakers(ctx)
			if service.IsFatal(err) {
				return err
			}
			if err != nil {
				log.Printf("failed to load takers: %s", err)
			}
		}
	}
}

func (s *Server) sendThanks(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ok, err := s.Service.SendNextThanks(ctx)
			if service.IsFatal(err) {
				return err
			}
			if err != nil {
				log.Printf("failed to send thanks: %s", err)
				continue
			}

			if ok {
				continue
			}

			log.Printf("sender will sleep for %s", s.PollInterval.String())

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
