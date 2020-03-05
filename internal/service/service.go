package service

import (
	"context"

	"github.com/vestatus/toggl/internal/logger"

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
	takerAPI     TakerAPI
	emailService EmailSender

	takerQueue TakerQueue
	sentThanks IDSet
}

func New(takerAPI TakerAPI, emailService EmailSender, takerQueue TakerQueue, sentThanks IDSet) *Service {
	return &Service{
		takerAPI:     takerAPI,
		emailService: emailService,
		takerQueue:   takerQueue,
		sentThanks:   sentThanks,
	}
}

func shouldSendThanks(taker *Taker) bool {
	const minPercent = 80

	return !taker.Demo && taker.Percent >= minPercent
}

func (s *Service) LoadTakers(ctx context.Context) (e error) {
	log := logger.FromContext(ctx)
	log.Debug("loading new takers")

	takers, err := s.takerAPI.ListTakers(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list takers")
	}

	loadedTakersCounter := 0
	defer func() { log.Debugf("loaded %v takers", loadedTakersCounter) }()

	for i := range takers {
		if !shouldSendThanks(&takers[i]) {
			continue
		}

		msgSent, err := s.sentThanks.Contains(ctx, takers[i].ID)
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to check for id in set"))
		}

		if msgSent {
			continue
		}

		err = s.takerQueue.Push(ctx, &takers[i])
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to push taker to the queue"))
		}

		err = s.sentThanks.Add(ctx, takers[i].ID)
		if err != nil {
			return Fatal(errors.Wrap(err, "failed to add taker to the queue"))
		}

		loadedTakersCounter++
	}

	return nil
}

func (s *Service) SendNextThanks(ctx context.Context) (ok bool, e error) {
	taker, err := s.takerQueue.Pop(ctx)
	if errors.Cause(err) == ErrNoTakers {
		return false, nil
	}
	if err != nil {
		return false, Fatal(errors.Wrap(err, "failed to pop taker"))
	}

	/*
		Name, address, subject and body shouldn't be hardcoded, but un-hardcoding the body would
		normally require templates and maybe a store for these templates, so I'll just leave it as is.
	*/
	err = s.emailService.SendEmail(EmailAddress{
		Name:    taker.Name,
		Address: taker.Email,
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
