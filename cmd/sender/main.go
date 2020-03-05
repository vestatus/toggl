package main

import (
	"context"
	"log"
	"net/http"
	"time"
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

	client, err := takers.NewClient(&http.Client{}, config.TakersAPI)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := client.Authenticate(ctx, config.Email, config.Password)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(token)
}
