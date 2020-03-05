package service

import (
	"context"
	"log"

	"github.com/pkg/errors"
)

type TakerAPI interface {
	Authenticate(ctx context.Context) error
	ListTakers(ctx context.Context) ([]Taker, error)
}

var ErrNoTakers = errors.New("no takers in queue")

type TakerQueue interface {
	Push(taker *Taker) error
	// A TakerQueue should return ErrNoTakers if no takers are available
	Pop() (*Taker, error)
}

type Service struct {
	TakerAPI     TakerAPI
	EmailService EmailSender
	TakerQueue   TakerQueue
}

func (s *Service) LoadTakers(ctx context.Context) error {
	log.Print("loading new takers")

	takers, err := s.TakerAPI.ListTakers(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list takers")
	}

	for i := range takers {
		err = s.TakerQueue.Push(&takers[i])
		if err != nil {
			return errors.Wrap(err, "failed to push taker to the queue")
		}
	}

	return nil
}

func (s *Service) SendNextThanks(_ context.Context) (ok bool, e error) {
	taker, err := s.TakerQueue.Pop()
	if errors.Cause(err) == ErrNoTakers {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to pop taker")
	}

	err = s.EmailService.SendEmail(EmailAddress{
		Name:    taker.Name,
		Address: taker.ContactEmail,
	}, EmailAddress{
		Name:    "Toggl Hire",
		Address: "hello@hire.toggl.com",
	}, Email{
		Subject: "Thank you",
		Body:    "Thank you for applying via Toggl Hire.",
	})
	if err != nil {
		return false, errors.Wrap(err, "failed to send email")
	}

	return true, nil
}
