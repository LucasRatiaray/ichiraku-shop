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

var _ usecase.OrderRepository = (*OrderRepository)(nil)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreatePaid(ctx context.Context, o *domain.Order, p *domain.Payment, basketID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := insertOrder(ctx, tx, o); err != nil {
		return err
	}

	p.OrderID = o.ID
	err = tx.QueryRowContext(ctx,
		`INSERT INTO payments (order_id, amount, card_last4) VALUES ($1, $2, $3) RETURNING id, created_at`,
		p.OrderID, p.Amount, p.CardLast4,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM basket_items WHERE basket_id = $1`, basketID); err != nil {
		return fmt.Errorf("clear basket: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, o *domain.Order) error {
	err := tx.QueryRowContext(ctx,
		`INSERT INTO orders (business_id, user_id, status, cancel_reason, total_ttc)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		o.BusinessID, o.UserID, o.Status, o.CancelReason, o.TotalTTC,
	).Scan(&o.ID, &o.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert order: %w", err)
	}

	for _, it := range o.Items {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, product_business_id, name, unit_price_ttc, quantity)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			o.ID, it.ProductID, it.ProductBusinessID, it.Name, it.UnitPriceTTC, it.Quantity)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}
	return nil
}
