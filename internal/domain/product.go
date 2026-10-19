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

const VATRate = 2000

func (p Product) PriceTTC() int64 {
	return (p.PriceHT*(10000+VATRate) + 5000) / 10000
}
