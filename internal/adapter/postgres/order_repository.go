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
	tx, err := r.db.BeginTx(ctx, nil) // commence une transaction pour créer la commande, le paiement et vider le panier
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // annule la transaction si une erreur survient avant le commit

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

func (r *OrderRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Order, error) {
	return r.list(ctx, "WHERE o.user_id = $1", userID)
}

func (r *OrderRepository) list(ctx context.Context, where string, args ...any) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT o.id, o.business_id, o.user_id, u.email, o.status, o.cancel_reason, o.total_ttc, o.created_at
	FROM orders o
	JOIN users u ON u.id = o.user_id `+where+`
	ORDER BY o.created_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := []domain.Order{}
	index := make(map[int64]int)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.BusinessID, &o.UserID, &o.UserEmail, &o.Status, &o.CancelReason, &o.TotalTTC, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		index[o.ID] = len(orders)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	itemRows, err := r.db.QueryContext(ctx, `
		SELECT oi.order_id, oi.product_id, oi.product_business_id, oi.name, oi.unit_price_ttc, oi.quantity
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id `+where+`
		ORDER BY oi.name`, args...)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var orderID int64
		var it domain.OrderItem
		if err := itemRows.Scan(&orderID, &it.ProductID, &it.ProductBusinessID, &it.Name, &it.UnitPriceTTC, &it.Quantity); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		i := index[orderID]
		orders[i].Items = append(orders[i].Items, it)
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	return r.list(ctx, "")
}

func (r *OrderRepository) FindByBusinessID(ctx context.Context, businessID string) (domain.Order, error) {
	orders, err := r.list(ctx, "WHERE o.business_id = $1", businessID)
	if err != nil {
		return domain.Order{}, err
	}
	if len(orders) == 0 {
		return domain.Order{}, domain.ErrNotFound
	}
	return orders[0], nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, from, to domain.OrderStatus, cancelReason string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, cancel_reason = $2 WHERE id = $3 AND status = $4`,
		to, cancelReason, id, from)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if n == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := insertOrder(ctx, tx, o); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
