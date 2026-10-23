package domain

import (
	"errors"
	"testing"
	"time"
)

func TestLuhnValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"carte de test valide", "4242424242424242", true},
		{"dernier chiffre modifié", "4242424242424241", false},
		{"exemple classique de 11 chiffres", "79927398713", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := luhnValid(tt.number); got != tt.want {
				t.Errorf("luhnValid(%q) = %v, attendu %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestValidateCard(t *testing.T) {
	now := time.Date(2026, time.October, 19, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		number string
		expiry string
		cvc    string
		want   error
	}{
		{"carte valide", "4242424242424242", "12/30", "123", nil},
		{"numéro avec espaces", "4242 4242 4242 4242", "12/30", "123", nil},
		{"mauvais Luhn", "4242424242424241", "12/30", "123", ErrInvalidCardNumber},
		{"numéro avec des lettres", "4242abcd42424242", "12/30", "123", ErrInvalidCardNumber},
		{"expiration le mois précédent", "4242424242424242", "09/26", "123", ErrCardExpired},
		{"expiration le mois en cours", "4242424242424242", "10/26", "123", nil},
		{"mois 13", "4242424242424242", "13/30", "123", ErrInvalidExpiry},
		{"format sans séparateur", "4242424242424242", "1226", "123", ErrInvalidExpiry},
		{"CVC trop court", "4242424242424242", "12/30", "12", ErrInvalidCVC},
		{"CVC avec une lettre", "4242424242424242", "12/30", "12a", ErrInvalidCVC},
		{"CVC de 4 chiffres", "4242424242424242", "12/30", "1234", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateCard(tt.number, tt.expiry, tt.cvc, now)
			if !errors.Is(got, tt.want) {
				t.Errorf("ValidateCard(%q, %q, %q) = %v, attendu %v", tt.number, tt.expiry, tt.cvc, got, tt.want)
			}
		})
	}
}
