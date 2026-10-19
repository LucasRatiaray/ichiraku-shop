// créer, changer statut, lister
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

type OrderService struct {
	orders   OrderRepository
	baskets  BasketRepository
	users    UserRepository
	products ProductRepository
}

func NewOrderService(orders OrderRepository, baskets BasketRepository, users UserRepository, products ProductRepository) *OrderService {
	return &OrderService{orders: orders, baskets: baskets, users: users, products: products}
}

type OrderLine struct {
	ProductID string
	Quantity  int
}

func (s *OrderService) Pay(ctx context.Context, userID int64, cardNumber, expiry, cvc string) (domain.Order, error) {
	if err := domain.ValidateCard(cardNumber, expiry, cvc, time.Now()); err != nil {
		return domain.Order{}, err
	}

	b, err := s.baskets.FindByUserID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Order{}, domain.ErrEmptyBasket
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("load basket: %w", err)
	}

	o, err := domain.NewOrderFromBasket(b, domain.OrderPaid)
	if err != nil {
		return domain.Order{}, err
	}

	for attempt := 0; attempt < maxIDAttempts; attempt++ {
		id, err := domain.NewBusinessID(domain.PrefixOrder)
		if err != nil {
			return domain.Order{}, fmt.Errorf("generate order id: %w", err)
		}
		o.BusinessID = id

		p := domain.Payment{
			Amount:    o.TotalTTC,
			CardLast4: domain.CardLast4(cardNumber),
		}

		err = s.orders.CreatePaid(ctx, &o, &p, b.ID)
		if err == nil {
			return o, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return domain.Order{}, fmt.Errorf("create order: %w", err)
		}
	}
	return domain.Order{}, fmt.Errorf("generate order id: %w", domain.ErrConflict)
}

func (s *OrderService) ListMine(ctx context.Context, userID int64) ([]domain.Order, error) {
	orders, err := s.orders.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) ListAll(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.orders.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) ChangeStatus(ctx context.Context, orderID string, to domain.OrderStatus, reason string) (domain.Order, error) {
	o, err := s.orders.FindByBusinessID(ctx, strings.ToUpper(strings.TrimSpace(orderID)))
	if err != nil {
		return domain.Order{}, err
	}

	from := o.Status
	if err := o.ChangeStatus(to, reason); err != nil {
		return domain.Order{}, err
	}

	if err := s.orders.UpdateStatus(ctx, o.ID, from, o.Status, o.CancelReason); err != nil {
		return domain.Order{}, err
	}
	return o, nil
}

func (s *OrderService) CreateForUser(ctx context.Context, email string, lines []OrderLine) (domain.Order, error) {
	if len(lines) == 0 {
		return domain.Order{}, domain.ErrInvalidInput
	}

	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return domain.Order{}, err
	}

	quantities := make(map[string]int)
	var order []string
	for _, l := range lines {
		id := strings.ToUpper(strings.TrimSpace(l.ProductID))
		if _, seen := quantities[id]; !seen {
			order = append(order, id)
		}
		quantities[id] += l.Quantity
	}

	o := domain.Order{UserID: user.ID, UserEmail: user.Email, Status: domain.OrderPending}
	for _, id := range order {
		qty := quantities[id]
		if qty < 1 || qty > maxQuantity {
			return domain.Order{}, domain.ErrInvalidInput
		}
		p, err := s.products.FindByBusinessID(ctx, id)
		if err != nil {
			return domain.Order{}, err
		}
		item := domain.OrderItem{
			ProductID:         p.ID,
			ProductBusinessID: p.BusinessID,
			Name:              p.Name,
			UnitPriceTTC:      p.PriceTTC(),
			Quantity:          qty,
		}
		o.Items = append(o.Items, item)
		o.TotalTTC += item.TotalTTC()
	}
	o.TotalTTC += domain.ShippingCost

	for attempt := 0; attempt < maxIDAttempts; attempt++ {
		bid, err := domain.NewBusinessID(domain.PrefixOrder)
		if err != nil {
			return domain.Order{}, fmt.Errorf("generate order id: %w", err)
		}
		o.BusinessID = bid

		err = s.orders.Create(ctx, &o)
		if err == nil {
			return o, nil
		}
		if !errors.Is(err, domain.ErrConflict) {
			return domain.Order{}, fmt.Errorf("create order: %w", err)
		}
	}
	return domain.Order{}, fmt.Errorf("generate order id: %w", domain.ErrConflict)
}
