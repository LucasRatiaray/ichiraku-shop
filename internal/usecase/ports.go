// Interfaces UserRepository, ProductRepository, CartRepository, OrderRepository...
package usecase

import (
	"context"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	List(ctx context.Context) ([]domain.User, error)
}

type VerificationCodeRepository interface {
	Create(ctx context.Context, code *domain.VerificationCode) error
	FindValid(ctx context.Context, userID int64, code string, purpose domain.CodePurpose) (domain.VerificationCode, error)
	MarkUsed(ctx context.Context, id int64) error
}
