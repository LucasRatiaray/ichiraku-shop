package usecase

import (
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

const (
	MaxCodeAttempts  = 5
	MaxLoginAttempts = 5
	LockDuration     = 15 * time.Minute
)

func isLocked(user domain.User, now time.Time) bool {
	return user.LockedUntil != nil && user.LockedUntil.After(now)
}
