package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	query := `INSERT INTO products (business_id, name, description, category, price_ht) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, p.BusinessID, p.Name, p.Description, p.Category, p.PriceHT).Scan(&p.ID, &p.CreatedAt)
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
	query := `SELECT id, business_id, name, description, category, price_ht, vat_rate, created_at FROM products WHERE business_id = $1`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.VATRate, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, fmt.Errorf("find product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) Search(ctx context.Context, f domain.ProductFilter) ([]domain.Product, error) {
	var conds []string
	var args []any

	add := func(cond string, arg any) {
		args = append(args, arg)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	ttc := fmt.Sprintf("((price_ht * %d + 5000) / 10000)", 10000+domain.VATRate)

	if f.Query != "" {
		add("(name ILIKE $%[1]d OR description ILIKE $%[1]d)", "%"+f.Query+"%")
	}
	if f.Category != "" {
		add("category ILIKE $%d", f.Category)
	}
	if f.MinPriceHT > 0 {
		add("price_ht >= $%d", f.MinPriceHT)
	}
	if f.MaxPriceHT > 0 {
		add("price_ht <= $%d", f.MaxPriceHT)
	}
	if f.MinPriceTTC > 0 {
		add(ttc+" >= $%d", f.MinPriceTTC)
	}
	if f.MaxPriceTTC > 0 {
		add(ttc+" <= $%d", f.MaxPriceTTC)
	}

	query := `SELECT id, business_id, name, description, category, price_ht, created_at FROM products`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY name LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}
	return products, nil
}
