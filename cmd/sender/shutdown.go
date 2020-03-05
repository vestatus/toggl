package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vestatus/toggl/internal/logger"
)

const maxGracePeriod = 6 * time.Second

type errSignal struct {
	Signal os.Signal
}

func (e errSignal) Error() string {
	return fmt.Sprintf("got signal %s", e.Signal)
}

func sigTrap(ctx context.Context) func() error {
	return func() error {
		trap := make(chan os.Signal, 1)

		signal.Notify(trap, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-trap:
			// in case the service fails to kill itself
			time.AfterFunc(maxGracePeriod, func() {
				logger.FromContext(ctx).Fatal("service failed to shut down gracefully")
			})

			return errSignal{Signal: sig}
		}
	}
}
