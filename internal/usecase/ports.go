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

type Mailer interface {
	SendCode(ctx context.Context, email string, code string, purpose domain.CodePurpose) error
}

type SessionRepository interface {
	Create(ctx context.Context, code *domain.Session) error
	FindValid(ctx context.Context, tokenHash string) (domain.Session, error)
	Delete(ctx context.Context, tokenHash string) error
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	FindByBusinessID(ctx context.Context, businessID string) (domain.Product, error)
	Search(ctx context.Context, f domain.ProductFilter) ([]domain.Product, error)
}

type BasketRepository interface {
	FindByUserID(ctx context.Context, userID int64) (domain.Basket, error)
	Create(ctx context.Context, b *domain.Basket) error
	SetItem(ctx context.Context, basketID, productID int64, quantity int) error
	RemoveItem(ctx context.Context, basketID, productID int64) error
}

type OrderRepository interface {
	CreatePaid(ctx context.Context, o *domain.Order, p *domain.Payment, basketID int64) error
	ListByUser(ctx context.Context, userID int64) ([]domain.Order, error)
	ListAll(ctx context.Context) ([]domain.Order, error)
	FindByBusinessID(ctx context.Context, businessID string) (domain.Order, error)
	UpdateStatus(ctx context.Context, id int64, from, to domain.OrderStatus, cancelReason string) error
	Create(ctx context.Context, o *domain.Order) error
}
