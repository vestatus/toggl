package main

import (
	"context"
	"net/http"
	"os"

	"github.com/vestatus/toggl/internal/email"
	"github.com/vestatus/toggl/internal/logger"
	"github.com/vestatus/toggl/internal/server"
	"github.com/vestatus/toggl/internal/service"
	"github.com/vestatus/toggl/internal/store"
	"github.com/vestatus/toggl/internal/takers"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"

	"golang.org/x/sync/errgroup"
)

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

	ctx := logger.WithLogger(context.Background(), log)

	err = run(ctx, config)
	if err != nil {
		log.WithError(err).Fatal("service terminated with error")
	}
}

func run(ctx context.Context, config *Config) error {
	log := logger.FromContext(ctx)

	client, err := takers.NewClient(&http.Client{}, config.TakersAPI, config.Email, config.Password)
	if err != nil {
		return errors.Wrap(err, "failed to create takers API client")
	}

	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.RedisAddr,
	})
	if err := redisClient.Ping().Err(); err != nil {
		return errors.Wrap(err, "failed to ping redis")
	}

	DB := store.NewRedis(redisClient)
	svc := service.New(client, &email.LogSender{Log: log}, DB, DB)

	srv := server.New(config.Server, svc)

	log.Info("starting service")

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return srv.Run(ctx)
	})
	eg.Go(sigTrap(ctx))

	err = eg.Wait()
	if sig, ok := err.(errSignal); ok {
		log.WithField("signal", sig.Signal).Info("service terminated normally due to signal")
		return nil
	}

	return err
}
