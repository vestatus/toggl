package service

// Email sending interface
// This should definitely take context.Context :(
type EmailSender interface {
	SendEmail(to, from EmailAddress, email Email) error
}

type EmailAddress struct {
	Name    string
	Address string
}

type Email struct {
	Subject string
	Body    string
}
