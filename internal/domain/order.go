package domain

import (
	"errors"
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
