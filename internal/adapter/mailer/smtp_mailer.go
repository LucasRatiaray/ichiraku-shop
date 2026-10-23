package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

var _ usecase.Mailer = (*SMTPMailer)(nil)

const smtpTimeout = 10 * time.Second

type SMTPMailer struct {
	host     string
	port     string
	from     string
	username string
	password string
}

func NewSMTPMailer(host, port, from, username, password string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, from: from, username: username, password: password}
}

func (m *SMTPMailer) SendCode(ctx context.Context, email string, code string, purpose domain.CodePurpose) error {
	subject := "Ichiraku Shop - Réinitialisation du mot de passe"
	body := "Bonjour,\n\nVoici ton code de réinitialisation : " + code + "\n\nIl est valable 15 minutes.\n"
	if purpose == domain.PurposeConfirm {
		subject = "Ichiraku Shop - Confirme ton compte"
		body = "Bonjour,\n\nVoici ton code de confirmation : " + code + "\n\nIl est valable 15 minutes.\n"
	}

	headers := []string{
		"From: " + m.from,
		"To: " + email,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n")

	return m.send(ctx, email, []byte(msg))
}

func (m *SMTPMailer) send(ctx context.Context, to string, msg []byte) error {
	dialer := net.Dialer{Timeout: smtpTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(m.host, m.port))
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("smtp deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if m.username != "" {
		if err := client.Auth(smtp.PlainAuth("", m.username, m.password, m.host)); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}
