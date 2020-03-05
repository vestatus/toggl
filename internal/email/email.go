package email

import (
	"toggl/internal/logger"
	"toggl/internal/service"
)

type LogSender struct {
	Log logger.Logger
}

func (s *LogSender) SendEmail(to, from service.EmailAddress, email service.Email) error {
	s.Log.Infof("to: %#+v\n", to)
	s.Log.Infof("from: %#+v\n", from)
	s.Log.Infof("email: %#+v\n", email)
	return nil
}
