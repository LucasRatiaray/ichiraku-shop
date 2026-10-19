package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ usecase.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (business_id, name, description, category, price_ht, vat_rate)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		p.BusinessID, p.Name, p.Description, p.Category, p.PriceHT, p.VATRate,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (r *ProductRepository) FindByBusinessID(ctx context.Context, businessID string) (domain.Product, error) {
	query := `SELECT id, business_id, name, description, category, price_ht, vat_rate, created_at
	          FROM products WHERE business_id = $1`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.VATRate, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, fmt.Errorf("find product: %w", err)
	}
	return p, nil
}
