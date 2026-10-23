package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ usecase.UserRepository = (*UserRepository)(nil) // permet de vérifier que UserRepository implémente l'interface usecase.UserRepository

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, role, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.Role, user.Status).Scan(&user.ID, &user.CreatedAt) // execute la requete, renvoie une ligne, copie les valeurs de la ligne dans les variables
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT id, email, password_hash, role, status, token_version, failed_logins, locked_until, created_at FROM users WHERE email = $1`
	var user domain.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.TokenVersion, &user.FailedLogins, &user.LockedUntil, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	query := `SELECT id, email, password_hash, role, status, token_version, failed_logins, locked_until, created_at FROM users WHERE id = $1`
	var user domain.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.TokenVersion, &user.FailedLogins, &user.LockedUntil, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	query := `UPDATE users SET email = $1, password_hash = $2, role = $3, status = $4 WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.PasswordHash, user.Role, user.Status, user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserRepository) IncrementTokenVersion(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET token_version = token_version + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("increment token version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("increment token version: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserRepository) RecordFailedLogin(ctx context.Context, id int64, maxAttempts int, lockedUntil time.Time) error {
	query := `UPDATE users SET
		failed_logins = CASE WHEN failed_logins + 1 >= $2 THEN 0 ELSE failed_logins + 1 END,
		locked_until = CASE WHEN failed_logins + 1 >= $2 THEN $3 ELSE locked_until END
		WHERE id = $1`

	if _, err := r.db.ExecContext(ctx, query, id, maxAttempts, lockedUntil); err != nil {
		return fmt.Errorf("record failed login: %w", err)
	}
	return nil
}

func (r *UserRepository) ResetFailedLogins(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET failed_logins = 0, locked_until = NULL WHERE id = $1`, id); err != nil {
		return fmt.Errorf("reset failed logins: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, password_hash, role, status, token_version, failed_logins, locked_until, created_at FROM users ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.TokenVersion, &user.FailedLogins, &user.LockedUntil, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // code d'erreur pour violation de contrainte de clé étrangère
			return domain.ErrConflict
		}
		return fmt.Errorf("delete user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
