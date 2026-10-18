// login, register, reset, confirm
package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (domain.User, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		return domain.User{}, domain.ErrInvalidInput
	}

	if len(password) < 8 || len(password) > 72 {
		return domain.User{}, domain.ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	u := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		Status:       domain.StatusPending,
	}

	if err := s.users.Create(ctx, &u); err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	return u, nil
}
