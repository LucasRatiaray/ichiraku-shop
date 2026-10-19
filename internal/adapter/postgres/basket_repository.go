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

var _ usecase.BasketRepository = (*BasketRepository)(nil)

type BasketRepository struct {
	db *sql.DB
}

func NewBasketRepository(db *sql.DB) *BasketRepository {
	return &BasketRepository{db: db}
}

func (r *BasketRepository) FindByUserID(ctx context.Context, userID int64) (domain.Basket, error) {
	var b domain.Basket
	err := r.db.QueryRowContext(ctx,
		`SELECT id, business_id, user_id, created_at FROM baskets WHERE user_id = $1`, userID,
	).Scan(&b.ID, &b.BusinessID, &b.UserID, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Basket{}, domain.ErrNotFound
		}
		return domain.Basket{}, fmt.Errorf("find basket: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.business_id, p.name, p.description, p.category, p.price_ht, p.created_at, bi.quantity
		FROM basket_items bi
		JOIN products p ON p.id = bi.product_id
		WHERE bi.basket_id = $1
		ORDER BY p.name`, b.ID)
	if err != nil {
		return domain.Basket{}, fmt.Errorf("find basket items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.BasketItem
		p := &item.Product
		if err := rows.Scan(&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.CreatedAt, &item.Quantity); err != nil {
			return domain.Basket{}, fmt.Errorf("scan basket item: %w", err)
		}
		b.Items = append(b.Items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Basket{}, fmt.Errorf("iterate basket items: %w", err)
	}
	return b, nil
}

func (r *BasketRepository) Create(ctx context.Context, b *domain.Basket) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO baskets (business_id, user_id) VALUES ($1, $2) RETURNING id, created_at`,
		b.BusinessID, b.UserID,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("create basket: %w", err)
	}
	return nil
}

func (r *BasketRepository) SetItem(ctx context.Context, basketID, productID int64, quantity int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO basket_items (basket_id, product_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (basket_id, product_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		basketID, productID, quantity)
	if err != nil {
		return fmt.Errorf("set basket item: %w", err)
	}
	return nil
}

func (r *BasketRepository) RemoveItem(ctx context.Context, basketID, productID int64) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM basket_items WHERE basket_id = $1 AND product_id = $2`, basketID, productID)
	if err != nil {
		return fmt.Errorf("remove basket item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove basket item: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
