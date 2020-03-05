package main

import (
	"context"
	"log"
	"net/http"
	"toggl/internal/email"
	"toggl/internal/queue"
	"toggl/internal/service"
	"toggl/internal/takers"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Email     string `envconfig:"EMAIL"`
	Password  string `envconfig:"PASSWORD"`
	TakersAPI string `envconfig:"TAKERS_API"`
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

	err = svc.LoadTakers(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	_, err = svc.SendNextThanks(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
