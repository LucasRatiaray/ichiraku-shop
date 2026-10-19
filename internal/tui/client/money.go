package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInvalidPrice = errors.New("prix invalide (ex : 12,50)")

func parseEuros(s string) (int64, error) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	if s == "" {
		return 0, nil
	}

	whole, frac, _ := strings.Cut(s, ".")
	if len(frac) > 2 {
		return 0, errInvalidPrice
	}
	for len(frac) < 2 {
		frac += "0"
	}

	euros, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || euros < 0 {
		return 0, errInvalidPrice
	}
	cents, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || cents < 0 {
		return 0, errInvalidPrice
	}
	return euros*100 + cents, nil
}

func formatEuros(cents int64) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

func validatePrice(s string) error {
	_, err := parseEuros(s)
	return err
}
