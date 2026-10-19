package domain

import "time"

type Payment struct {
	ID        int64
	OrderID   int64
	Amount    int64
	CardLast4 string
	CreatedAt time.Time
}

func CardLast4(number string) string {
	n := NormalizeCardNumber(number)
	if len(n) < 4 {
		return ""
	}
	return n[len(n)-4:]
}