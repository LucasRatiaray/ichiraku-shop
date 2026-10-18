package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

var _ usecase.VerificationCodeRepository = (*VerificationCodeRepository)(nil)

type VerificationCodeRepository struct {
	db *sql.DB
}

func NewVerificationCodeRepository(db *sql.DB) *VerificationCodeRepository {
	return &VerificationCodeRepository{db: db}
}

func (r *VerificationCodeRepository) Create(ctx context.Context, vc *domain.VerificationCode) error {
	query := `INSERT INTO verification_codes (user_id, code, purpose, expires_at) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRowContext(ctx, query, vc.UserID, vc.Code, vc.Purpose, vc.ExpiresAt).Scan(&vc.ID)
	if err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

func (r *VerificationCodeRepository) FindValid(ctx context.Context, userID int64, code string, purpose domain.CodePurpose) (domain.VerificationCode, error) {
	query := `SELECT id, user_id, code, purpose, expires_at, used_at FROM verification_codes WHERE user_id = $1 AND code = $2 AND purpose = $3 AND used_at IS NULL AND expires_at > now() ORDER BY id DESC LIMIT 1`
	var vc domain.VerificationCode

	err := r.db.QueryRowContext(ctx, query, userID, code, purpose).Scan(&vc.ID, &vc.UserID, &vc.Code, &vc.Purpose, &vc.ExpiresAt, &vc.UsedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.VerificationCode{}, domain.ErrNotFound
		}
		return domain.VerificationCode{}, fmt.Errorf("find valid code: %w", err)
	}
	return vc, nil
}

func (r *VerificationCodeRepository) MarkUsed(ctx context.Context, id int64) error {
	query := `UPDATE verification_codes SET used_at = now() WHERE id = $1 AND used_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}
