package server

import "time"

type Config struct {
	SendThanksTimeout time.Duration `envconfig:"SEND_THANKS_TIMEOUT"`
	UpdateInterval    time.Duration `envconfig:"UPDATE_INTERVAL"`
	PollInterval      time.Duration `envconfig:"POLL_INTERVAL"`
}
