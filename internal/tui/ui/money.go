package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidPrice = errors.New("invalid price (ex: 12,50)")

func ParseEuros(s string) (int64, error) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	if s == "" {
		return 0, nil
	}

	whole, frac, _ := strings.Cut(s, ".")
	if len(frac) > 2 {
		return 0, ErrInvalidPrice
	}
	for len(frac) < 2 {
		frac += "0"
	}

	euros, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || euros < 0 {
		return 0, ErrInvalidPrice
	}
	cents, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || cents < 0 {
		return 0, ErrInvalidPrice
	}
	return euros*100 + cents, nil
}

func FormatEuros(cents int64) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

func ValidatePrice(s string) error {
	_, err := ParseEuros(s)
	return err
}
