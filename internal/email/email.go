package email

import (
	"log"
	"toggl/internal/service"
)

type LogSender struct{}

func (s *LogSender) SendEmail(to, from service.EmailAddress, email service.Email) error {
	log.Printf("to: %#+v\n", to)
	log.Printf("from: %#+v\n", from)
	log.Printf("email: %#+v\n", email)
	return nil
}
