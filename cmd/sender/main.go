package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"toggl/internal/email"
	"toggl/internal/queue"
	"toggl/internal/server"
	"toggl/internal/service"
	"toggl/internal/takers"

	"golang.org/x/sync/errgroup"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Email     string `envconfig:"EMAIL"`
	Password  string `envconfig:"PASSWORD"`
	TakersAPI string `envconfig:"TAKERS_API"`
}

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
				log.Fatal("service failed to shut down gracefully")
			})

			return errSignal{Signal: sig}
		}
	}
}

func main() {
	var config Config

	err := envconfig.Process("sender", &config)
	if err != nil {
		log.Fatal(err)
	}

	client, err := takers.NewClient(&http.Client{}, config.TakersAPI, config.Email, config.Password)
	if err != nil {
		log.Fatal(err)
	}

	svc := &service.Service{
		TakerAPI:     client,
		EmailService: &email.LogSender{},
		TakerQueue:   queue.NewInmem(),
	}

	srv := server.Server{
		UpdateInterval: 10 * time.Second,
		Service:        svc,
		PollInterval:   5 * time.Second,
	}

	eg, ctx := errgroup.WithContext(context.TODO())
	eg.Go(func() error {
		return srv.Run(ctx)
	})
	eg.Go(sigTrap(ctx))

	err = eg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
