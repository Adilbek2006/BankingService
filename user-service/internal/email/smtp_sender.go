package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type Sender interface {
	Send(to, subject, body string) error
}

type SMTPSender struct {
	host               string
	port               string
	username           string
	password           string
	from               string
	useTLS             bool
	insecureSkipVerify bool
}

func NewSMTPSender(host, port, username, password, from string, useTLS, insecureSkipVerify bool) *SMTPSender {
	return &SMTPSender{
		host:               host,
		port:               port,
		username:           username,
		password:           password,
		from:               from,
		useTLS:             useTLS,
		insecureSkipVerify: insecureSkipVerify,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	if s.host == "" || s.port == "" {
		return fmt.Errorf("smtp host/port not configured")
	}

	addr := net.JoinHostPort(s.host, s.port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if s.useTLS {
		tlsConfig := &tls.Config{ServerName: s.host, InsecureSkipVerify: s.insecureSkipVerify}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(s.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	message := buildMessage(s.from, to, subject, body)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (s *NoopSender) Send(to, subject, body string) error {
	_ = to
	_ = subject
	_ = body
	return nil
}

func buildMessage(from, to, subject, body string) string {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
	}
	return strings.Join(headers, "\r\n") + body
}
