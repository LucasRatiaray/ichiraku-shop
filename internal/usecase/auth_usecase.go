// login, register, reset, confirm
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users    UserRepository
	codes    VerificationCodeRepository
	mailer   Mailer
	sessions SessionRepository
}

func NewAuthService(users UserRepository, codes VerificationCodeRepository, mailer Mailer, sessions SessionRepository) *AuthService {
	return &AuthService{
		users:    users,
		codes:    codes,
		mailer:   mailer,
		sessions: sessions,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (domain.User, error) {
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

	code, err := generateCode()
	if err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	vc := domain.VerificationCode{
		UserID:    u.ID,
		Code:      code,
		Purpose:   domain.PurposeConfirm,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.codes.Create(ctx, &vc); err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	if err := s.mailer.SendCode(ctx, u.Email, code, domain.PurposeConfirm); err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	return u, nil
}

func (s *AuthService) Confirm(ctx context.Context, email, code string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}

	if u.Status != domain.StatusPending {
		return domain.ErrInvalidCode
	}

	vc, err := s.codes.FindValid(ctx, u.ID, code, domain.PurposeConfirm)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}

	err = s.codes.MarkUsed(ctx, vc.ID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}

	u.Status = domain.StatusActive
	if err := s.users.Update(ctx, u); err != nil {
		return fmt.Errorf("confirm: %w", err)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	if u.Status == domain.StatusPending {
		return "", domain.ErrAccountNotConfirmed
	}

	if u.Status == domain.StatusDisabled {
		return "", domain.ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	sess := domain.Session{
		UserID:    u.ID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessions.Create(ctx, &sess); err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	return token, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	sess, err := s.sessions.FindValid(ctx, hashToken(token))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("authenticate: %w", err)
	}

	u, err := s.users.FindByID(ctx, sess.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("authenticate: %w", err)
	}

	if u.Status != domain.StatusActive {
		return domain.User{}, domain.ErrUnauthorized
	}

	return u, nil
}
