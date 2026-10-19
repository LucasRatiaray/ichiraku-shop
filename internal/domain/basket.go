package domain

import "time"

const ShippingCost int64 = 0

type BasketItem struct {
	Product  Product
	Quantity int
}

func (i BasketItem) TotalTTC() int64 {
	return i.Product.PriceTTC() * int64(i.Quantity)
}

type Basket struct {
	ID         int64
	BusinessID string
	UserID     int64
	Items      []BasketItem
	CreatedAt  time.Time
}

func (b Basket) TotalTTC() int64 {
	var total int64
	for _, item := range b.Items {
		total += item.TotalTTC()
	}
	return total + ShippingCost
}