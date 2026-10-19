package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidCardNumber = errors.New("invalid card number")
	ErrInvalidExpiry     = errors.New("invalid expiry date (MM/YY)")
	ErrCardExpired       = errors.New("card expired")
	ErrInvalidCVC        = errors.New("invalid CVC")
)

func NormalizeCardNumber(number string) string {
	return strings.ReplaceAll(strings.TrimSpace(number), " ", "")
}

func luhnValid(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseExpiry(expiry string) (month, year int, err error) {
	mm, yy, ok := strings.Cut(strings.TrimSpace(expiry), "/")
	if !ok || len(mm) != 2 || len(yy) != 2 {
		return 0, 0, ErrInvalidExpiry
	}
	month, err = strconv.Atoi(mm)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, ErrInvalidExpiry
	}
	y, err := strconv.Atoi(yy)
	if err != nil {
		return 0, 0, ErrInvalidExpiry
	}
	return month, 2000 + y, nil
}

func ValidateCard(number, expiry, cvc string, now time.Time) error {
	number = NormalizeCardNumber(number)
	if !isDigits(number) || len(number) < 13 || len(number) > 19 || !luhnValid(number) {
		return ErrInvalidCardNumber
	}

	month, year, err := parseExpiry(expiry)
	if err != nil {
		return err
	}
	if year*12+month < now.Year()*12+int(now.Month()) {
		return ErrCardExpired
	}

	cvc = strings.TrimSpace(cvc)
	if !isDigits(cvc) || len(cvc) < 3 || len(cvc) > 4 {
		return ErrInvalidCVC
	}
	return nil
}