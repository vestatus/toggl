package main

import (
	"context"
	"net/http"
	"os"
	"toggl/internal/email"
	"toggl/internal/logger"
	"toggl/internal/server"
	"toggl/internal/service"
	"toggl/internal/store"
	"toggl/internal/takers"

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

	DB := store.NewRedis(redisClient)
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
