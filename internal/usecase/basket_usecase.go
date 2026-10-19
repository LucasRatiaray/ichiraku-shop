package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

const maxQuantity = 99

type BasketService struct {
	baskets  BasketRepository
	products ProductRepository
}

func NewBasketService(baskets BasketRepository, products ProductRepository) *BasketService {
	return &BasketService{baskets: baskets, products: products}
}

func (s *BasketService) Get(ctx context.Context, userID int64) (domain.Basket, error) {
	for attempt := 0; attempt < maxIDAttempts; attempt++ {
		b, err := s.baskets.FindByUserID(ctx, userID)
		if err == nil {
			return b, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.Basket{}, fmt.Errorf("get basket: %w", err)
		}

		id, err := domain.NewBusinessID(domain.PrefixBasket) // génère un nouvel ID unique pour le panier
		if err != nil {
			return domain.Basket{}, fmt.Errorf("generate basket id: %w", err)
		}
		nb := domain.Basket{BusinessID: id, UserID: userID}
		err = s.baskets.Create(ctx, &nb)
		if err == nil {
			return nb, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return domain.Basket{}, fmt.Errorf("create basket: %w", err)
		}
	}
	return domain.Basket{}, fmt.Errorf("get basket: %w", domain.ErrConflict)
}

func (s *BasketService) AddItem(ctx context.Context, userID int64, productID string, quantity int) (domain.Basket, error) {
	if quantity < 1 {
		return domain.Basket{}, domain.ErrInvalidInput
	}
	product, b, err := s.load(ctx, userID, productID)
	if err != nil {
		return domain.Basket{}, err
	}

	newQty := quantity
	for _, item := range b.Items {
		if item.Product.ID == product.ID {
			newQty += item.Quantity
			break
		}
	}
	if newQty > maxQuantity {
		return domain.Basket{}, domain.ErrInvalidInput
	}

	if err := s.baskets.SetItem(ctx, b.ID, product.ID, newQty); err != nil {
		return domain.Basket{}, fmt.Errorf("add item: %w", err)
	}
	return s.baskets.FindByUserID(ctx, userID)
}

func (s *BasketService) SetQuantity(ctx context.Context, userID int64, productID string, quantity int) (domain.Basket, error) {
	if quantity < 0 || quantity > maxQuantity {
		return domain.Basket{}, domain.ErrInvalidInput
	}
	if quantity == 0 {
		return s.RemoveItem(ctx, userID, productID)
	}
	product, b, err := s.load(ctx, userID, productID)
	if err != nil {
		return domain.Basket{}, err
	}
	if err := s.baskets.SetItem(ctx, b.ID, product.ID, quantity); err != nil {
		return domain.Basket{}, fmt.Errorf("set quantity: %w", err)
	}
	return s.baskets.FindByUserID(ctx, userID)
}

func (s *BasketService) RemoveItem(ctx context.Context, userID int64, productID string) (domain.Basket, error) {
	product, b, err := s.load(ctx, userID, productID)
	if err != nil {
		return domain.Basket{}, err
	}
	if err := s.baskets.RemoveItem(ctx, b.ID, product.ID); err != nil {
		return domain.Basket{}, err
	}
	return s.baskets.FindByUserID(ctx, userID)
}

func (s *BasketService) load(ctx context.Context, userID int64, productID string) (domain.Product, domain.Basket, error) {
	productID = strings.ToUpper(strings.TrimSpace(productID))
	product, err := s.products.FindByBusinessID(ctx, productID)
	if err != nil {
		return domain.Product{}, domain.Basket{}, err
	}
	b, err := s.Get(ctx, userID)
	if err != nil {
		return domain.Product{}, domain.Basket{}, err
	}
	return product, b, nil
}
