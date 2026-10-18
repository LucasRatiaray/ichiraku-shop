package mailer

import (
	"context"
	"log"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

var _ usecase.Mailer = (*LogMailer)(nil)

type LogMailer struct{}

func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) SendCode(ctx context.Context, email string, code string, purpose domain.CodePurpose) error {
	log.Printf("[mail] to=%s purpose=%s code=%s", email, purpose, code)
	return nil
}
