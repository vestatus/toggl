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

	client, err := takers.NewClient(&http.Client{}, config.TakersAPI, config.Email, config.Password)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Authenticate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	takers, err := client.ListTakers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, taker := range takers {
		log.Printf("%#v\n", taker)
	}
}
