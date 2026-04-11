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
	baseURL  string
}

func NewSender(from, password, host, port string, baseURL string) *Sender {
	return &Sender{from: from, password: password, host: host, port: port, baseURL: baseURL}
}

func (s *Sender) SendConfirmation(to, token string) error {
	subject := "Confirm your subscription"
	body := fmt.Sprintf("Please click this link for confirmation your subscription:  %s/api/confirm/%s", s.baseURL, token)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", s.from, to, subject, body)
	auth := smtp.PlainAuth("", s.from, s.password, s.host)

	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, []byte(message))
}

func (s *Sender) SendNotification(to, repoName, tag, unsubToken string) error {
	subject := fmt.Sprintf("New release %s for %s", tag, repoName)
	body := fmt.Sprintf("Repository %s has a new release: %s\n\nRelease page: https://github.com/%s/releases/tag/%s\n\nUnsubscribe: %s/api/unsubscribe/%s",
		repoName, tag, repoName, tag, s.baseURL, unsubToken)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", s.from, to, subject, body)
	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, []byte(message))
}
