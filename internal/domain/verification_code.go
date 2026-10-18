package domain

import "time"

type CodePurpose string

const (
	PurposeConfirm CodePurpose = "confirm"
	PurposeReset   CodePurpose = "reset"
)

type VerificationCode struct {
	ID        int64
	UserID    int64
	Code      string
	Purpose   CodePurpose
	ExpiresAt time.Time
	UsedAt    *time.Time
}
