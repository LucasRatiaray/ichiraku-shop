package domain

import "testing"

func TestPriceTTC(t *testing.T) {
	tests := []struct {
		name    string
		priceHT int64
		want    int64
	}{
		{"1000 centimes HT", 1000, 1200},
		{"700 centimes HT", 700, 840},
		{"799 centimes HT avec arrondi", 799, 959},
		{"prix nul", 0, 0},
		{"1 centime HT", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Product{PriceHT: tt.priceHT}).PriceTTC(); got != tt.want {
				t.Errorf("PriceTTC(%d) = %d, attendu %d", tt.priceHT, got, tt.want)
			}
		})
	}
}
