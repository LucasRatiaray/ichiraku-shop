// recherche, création admin
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

const maxIDAttempts = 3

type ProductService struct {
	products ProductRepository
}

func NewProductService(products ProductRepository) *ProductService {
	return &ProductService{
		products: products,
	}
}

func (s *ProductService) Create(ctx context.Context, name, description, category string, priceHT int64) (domain.Product, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	category = strings.TrimSpace(category)

	if name == "" || category == "" || priceHT <= 0 {
		return domain.Product{}, domain.ErrInvalidInput
	}

	p := domain.Product{
		Name:        name,
		Description: description,
		Category:    category,
		PriceHT:     priceHT,
		VATRate:     domain.VATRate,
	}

	for attempt := 0; attempt < maxIDAttempts; attempt++ {
		id, err := domain.NewBusinessID(domain.PrefixProduct)
		if err != nil {
			return domain.Product{}, fmt.Errorf("generate product id: %w", err)
		}
		p.BusinessID = id

		err = s.products.Create(ctx, &p)
		if err == nil {
			return p, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return domain.Product{}, fmt.Errorf("create product: %w", err)
		}
	}

	return domain.Product{}, fmt.Errorf("generate product id: %w", domain.ErrConflict)
}

func (s *ProductService) Search(ctx context.Context, f domain.ProductFilter) ([]domain.Product, error) {
	f.Query = strings.TrimSpace(f.Query)
	f.Category = strings.TrimSpace(f.Category)

	if f.MinPriceHT < 0 || f.MaxPriceHT < 0 || f.MinPriceTTC < 0 || f.MaxPriceTTC < 0 {
		return nil, domain.ErrInvalidInput
	}
	if f.MaxPriceHT > 0 && f.MinPriceHT > f.MaxPriceHT {
		return nil, domain.ErrInvalidInput
	}
	if f.MaxPriceTTC > 0 && f.MinPriceTTC > f.MaxPriceTTC {
		return nil, domain.ErrInvalidInput
	}

	products, err := s.products.Search(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	return products, nil
}
