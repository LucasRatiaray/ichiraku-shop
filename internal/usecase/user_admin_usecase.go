package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserAdminService struct {
	users UserRepository
}

func NewUserAdminService(users UserRepository) *UserAdminService {
	return &UserAdminService{users: users}
}

func validRole(r domain.Role) bool {
	return r == domain.RoleUser || r == domain.RoleAdmin
}

func (s *UserAdminService) List(ctx context.Context) ([]domain.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *UserAdminService) Create(ctx context.Context, email, password, role string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return domain.User{}, domain.ErrInvalidInput
	}
	if len(password) < 8 || len(password) > 72 {
		return domain.User{}, domain.ErrInvalidInput
	}
	r := domain.Role(role)
	if !validRole(r) {
		return domain.User{}, domain.ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{Email: email, PasswordHash: string(hash), Role: r, Status: domain.StatusActive}
	if err := s.users.Create(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *UserAdminService) Update(ctx context.Context, actorID, id int64, email, password, role, status string) (domain.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil || addr.Address != email {
			return domain.User{}, domain.ErrInvalidInput
		}
		user.Email = email
	}
	if password != "" {
		if len(password) < 8 || len(password) > 72 {
			return domain.User{}, domain.ErrInvalidInput
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return domain.User{}, fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = string(hash)
	}

	if role != "" {
		r := domain.Role(role)
		if !validRole(r) {
			return domain.User{}, domain.ErrInvalidInput
		}
		user.Role = r
	}
	if status != "" {
		st := domain.Status(status)
		if st != domain.StatusActive && st != domain.StatusDisabled {
			return domain.User{}, domain.ErrInvalidInput
		}
		user.Status = st
	}

	if id == actorID && (user.Role != domain.RoleAdmin || user.Status != domain.StatusActive) {
		return domain.User{}, domain.ErrSelfAction
	}

	if err := s.users.Update(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *UserAdminService) Delete(ctx context.Context, actorID, id int64) error {
	if id == actorID {
		return domain.ErrSelfAction
	}
	return s.users.Delete(ctx, id)
}
