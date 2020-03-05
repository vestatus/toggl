package service

import (
	"context"
	"toggl/internal/logger"

	"github.com/pkg/errors"
)

type TakerAPI interface {
	Authenticate(ctx context.Context) error
	ListTakers(ctx context.Context) ([]Taker, error)
}

var ErrNoTakers = errors.New("no takers in queue")

type TakerQueue interface {
	Push(ctx context.Context, taker *Taker) error
	// A TakerQueue should return ErrNoTakers if no takers are available
	Pop(ctx context.Context) (*Taker, error)
}

type IDSet interface {
	Add(ctx context.Context, id int) error
	Contains(ctx context.Context, id int) (bool, error)
}

type Service struct {
	TakerAPI     TakerAPI
	EmailService EmailSender
	TakerQueue   TakerQueue
	SentThanks   IDSet
}

func shouldSendThanks(taker *Taker) bool {
	const minPercent = 80

	return !taker.Demo && taker.Percent >= minPercent
}

func (s *Service) LoadTakers(ctx context.Context) error {
	logger.FromContext(ctx).Info("loading new takers")

	takers, err := s.TakerAPI.ListTakers(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list takers")
	}

	for i := range takers {
		if !shouldSendThanks(&takers[i]) {
			continue
		}

		msgSent, err := s.SentThanks.Contains(ctx, takers[i].ID)
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to check for id in set"))
		}

		if msgSent {
			continue
		}

		err = s.TakerQueue.Push(ctx, &takers[i])
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to push taker to the queue"))
		}

		err = s.SentThanks.Add(ctx, takers[i].ID)
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to add taker to the queue"))
		}
	}

	return nil
}

func (s *Service) SendNextThanks(ctx context.Context) (ok bool, e error) {
	taker, err := s.TakerQueue.Pop(ctx)
	if errors.Cause(err) == ErrNoTakers {
		return false, nil
	}
	if err != nil {
		return false, Fatal(errors.Wrap(err, "failed to pop taker"))
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
		return false, Fatal(errors.Wrap(err, "failed to send email"))
	}

	return true, nil
}
