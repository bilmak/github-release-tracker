package email

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	from     string
	password string
	host     string
	port     string
}

func NewSender(from, password, host, port string) *Sender {
	return &Sender{from: from, password: password, host: host, port: port}
}

func (s *Sender) SendConfirmation(to, token string) error {
	subject := "Confirm your subscription"
	body := fmt.Sprintf("Please click this link for confirmation your subscription:  http://localhost:8080/api/confirm?token=%s ", token)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, to, subject, body)
	auth := smtp.PlainAuth("", s.from, s.password, s.host)

	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, []byte(message))
}
