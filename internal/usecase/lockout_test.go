package usecase

import (
	"testing"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

func TestIsLocked(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Minute)
	past := now.Add(-time.Minute)

	tests := []struct {
		name  string
		until *time.Time
		want  bool
	}{
		{"jamais verrouille", nil, false},
		{"verrou dans le futur", &future, true},
		{"verrou expire", &past, false},
		{"verrou qui expire maintenant", &now, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLocked(domain.User{LockedUntil: tt.until}, now); got != tt.want {
				t.Errorf("isLocked = %v, attendu %v", got, tt.want)
			}
		})
	}
}
