package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"toggl/internal/db"
	"toggl/internal/email"
	"toggl/internal/logger"
	"toggl/internal/server"
	"toggl/internal/service"
	"toggl/internal/takers"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"

	"golang.org/x/sync/errgroup"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Email     string `envconfig:"EMAIL"`
	Password  string `envconfig:"PASSWORD"`
	TakersAPI string `envconfig:"TAKERS_API"`
	RedisAddr string `envconfig:"REDIS_ADDR"`
	LogLevel  string `envconfig:"LOG_LEVEL"`

	Server server.Config `envconfig:"SERVER"`
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
				logger.FromContext(ctx).Fatal("service failed to shut down gracefully")
			})

			return errSignal{Signal: sig}
		}
	}
}

func loadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process("sender", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	const defaultLogLevel = logrus.InfoLevel

	log := logrus.NewEntry(&logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.TextFormatter{},
		Level:     defaultLogLevel,
	})

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("config: %#v", config)

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		log.WithError(err).
			WithField("default level", defaultLogLevel).
			Error("failed to parse log level, using default")
	} else {
		log = logrus.NewEntry(&logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{},
			Level:     level,
		})
	}

	client, err := takers.NewClient(&http.Client{}, config.TakersAPI, config.Email, config.Password)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.RedisAddr,
	})
	if redisClient.Ping().Err() != nil {
		log.Fatal("redis unavailable")
	}

	DB := db.NewRedis(redisClient)
	svc := service.New(client, &email.LogSender{Log: log}, DB, DB)

	srv := server.New(config.Server, svc)

	ctx := logger.WithLogger(context.Background(), log)

	log.Info("starting service")

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return srv.Run(ctx)
	})
	eg.Go(sigTrap(ctx))

	err = eg.Wait()
	if sig, ok := err.(errSignal); ok {
		log.WithField("signal", sig.Signal).Info("service terminated normally due to signal")
		return
	}

	log.WithError(err).Error("service terminated with error")
}
