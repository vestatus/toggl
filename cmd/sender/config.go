package main

import (
	"toggl/internal/server"

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

func loadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process("sender", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
