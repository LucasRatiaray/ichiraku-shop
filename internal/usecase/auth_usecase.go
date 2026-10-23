// login, register, reset, confirm
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     UserRepository
	codes     VerificationCodeRepository
	mailer    Mailer
	jwtSecret []byte
}

func NewAuthService(users UserRepository, codes VerificationCodeRepository, mailer Mailer, jwtSecret []byte) *AuthService {
	return &AuthService{
		users:     users,
		codes:     codes,
		mailer:    mailer,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		return domain.User{}, domain.ErrInvalidInput
	}

	if err := validatePassword(password); err != nil {
		return domain.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	user := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		Status:       domain.StatusPending,
	}

	if err := s.users.Create(ctx, &user); err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	code, err := s.newCode(ctx, user.ID, domain.PurposeConfirm)
	if err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}

	if err := s.mailer.SendCode(ctx, user.Email, code, domain.PurposeConfirm); err != nil {
		log.Printf("register: send confirmation code: %v", err)
	}

	return user, nil
}

func (s *AuthService) newCode(ctx context.Context, userID int64, purpose domain.CodePurpose) (string, error) {
	code, err := generateCode()
	if err != nil {
		return "", err
	}

	vc := domain.VerificationCode{
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.codes.Create(ctx, &vc); err != nil {
		return "", err
	}
	return code, nil
}

func (s *AuthService) ResendConfirmation(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("resend confirmation: %w", err)
	}

	if user.Status != domain.StatusPending {
		return nil
	}

	code, err := s.newCode(ctx, user.ID, domain.PurposeConfirm)
	if err != nil {
		return fmt.Errorf("resend confirmation: %w", err)
	}

	if err := s.mailer.SendCode(ctx, user.Email, code, domain.PurposeConfirm); err != nil {
		return fmt.Errorf("resend confirmation: %w", err)
	}

	return nil
}

func (s *AuthService) Confirm(ctx context.Context, email, code string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}

	if user.Status != domain.StatusPending {
		return domain.ErrInvalidCode
	}

	vc, err := s.codes.FindValid(ctx, user.ID, code, domain.PurposeConfirm)
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

	user.Status = domain.StatusActive
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("confirm: %w", err)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	if user.Status == domain.StatusPending {
		return "", domain.ErrAccountNotConfirmed
	}

	if user.Status == domain.StatusDisabled {
		return "", domain.ErrInvalidCredentials
	}

	token, err := signToken(s.jwtSecret, user, time.Now())
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	return token, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	userID, version, err := parseToken(s.jwtSecret, token)
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}

	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("authenticate: %w", err)
	}

	if user.Status != domain.StatusActive || user.TokenVersion != version {
		return domain.User{}, domain.ErrUnauthorized
	}

	return user, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("request reset: %w", err)
	}

	if user.Status != domain.StatusActive {
		return nil
	}

	code, err := s.newCode(ctx, user.ID, domain.PurposeReset)
	if err != nil {
		return fmt.Errorf("request reset: %w", err)
	}

	if err := s.mailer.SendCode(ctx, user.Email, code, domain.PurposeReset); err != nil {
		return fmt.Errorf("request reset: %w", err)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	vc, err := s.codes.FindValid(ctx, user.ID, code, domain.PurposeReset)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	err = s.codes.MarkUsed(ctx, vc.ID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrInvalidCode
	}
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	user.PasswordHash = string(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := s.users.IncrementTokenVersion(ctx, user.ID); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	return nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	if err := s.users.IncrementTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

func (s *AuthService) EnsureAdmin(ctx context.Context, email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("ensure admin: %w", err)
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ensure admin: %w", err)
	}

	u := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleAdmin,
		Status:       domain.StatusActive,
	}
	if err := s.users.Create(ctx, &u); err != nil {
		return fmt.Errorf("ensure admin: %w", err)
	}
	return nil
}
