package usecase

import "github.com/LucasRatiaray/ichiraku-shop/internal/domain"

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return domain.ErrInvalidInput
	}
	return nil
}
