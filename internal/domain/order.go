package domain

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var ErrEmptyBasket = errors.New("basket is empty")

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderShipping  OrderStatus = "shipping"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID         int64
	ProductBusinessID string
	Name              string
	UnitPriceTTC      int64
	Quantity          int
}

func (i OrderItem) TotalTTC() int64 {
	return i.UnitPriceTTC * int64(i.Quantity)
}

type Order struct {
	ID           int64
	BusinessID   string
	UserID       int64
	Status       OrderStatus
	CancelReason string
	TotalTTC     int64
	Items        []OrderItem
	CreatedAt    time.Time
	UserEmail 	 string
}

func NewOrderFromBasket(b Basket, status OrderStatus) (Order, error) {
	if len(b.Items) == 0 {
		return Order{}, ErrEmptyBasket
	}

	items := make([]OrderItem, 0, len(b.Items))
	for _, it := range b.Items {
		items = append(items, OrderItem{
			ProductID:         it.Product.ID,
			ProductBusinessID: it.Product.BusinessID,
			Name:              it.Product.Name,
			UnitPriceTTC:      it.Product.PriceTTC(),
			Quantity:          it.Quantity,
		})
	}

	return Order{
		UserID:   b.UserID,
		Status:   status,
		TotalTTC: b.TotalTTC(),
		Items:    items,
	}, nil
}

var (
	ErrInvalidStatus        = errors.New("invalid status")
	ErrInvalidTransition    = errors.New("status transition not allowed")
	ErrCancelReasonRequired = errors.New("cancel reason is required")
)

var transitions = map[OrderStatus][]OrderStatus{
	OrderPending:   {OrderPaid, OrderCancelled},
	OrderPaid:      {OrderShipping, OrderCancelled},
	OrderShipping:  {OrderDelivered},
	OrderDelivered: {},
	OrderCancelled: {},
}

func (s OrderStatus) Valid() bool {
	_, ok := transitions[s]
	return ok
}

func (o *Order) ChangeStatus(to OrderStatus, reason string) error {
	if !to.Valid() {
		return ErrInvalidStatus
	}
	if !slices.Contains(transitions[o.Status], to) {
		return ErrInvalidTransition
	}
	if to == OrderCancelled {
		reason = strings.TrimSpace(reason)
		if reason == "" {
			return ErrCancelReasonRequired
		}
		o.CancelReason = reason
	}
	o.Status = to
	return nil
}
