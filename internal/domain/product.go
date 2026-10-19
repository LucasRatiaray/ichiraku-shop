package domain

import "time"

type Product struct {
	ID          int64
	BusinessID  string
	Name        string
	Description string
	Category    string
	PriceHT     int64
	VATRate     int64
	CreatedAt   time.Time
}

type ProductFilter struct {
	Query       string // nom ou description
	Category    string
	MinPriceHT  int64
	MaxPriceHT  int64
	MinPriceTTC int64
	MaxPriceTTC int64
}

const VATRate = 2000

func (p Product) PriceTTC() int64 {
	return (p.PriceHT*(10000+VATRate) + 5000) / 10000
}
