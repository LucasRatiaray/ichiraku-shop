// (génération PDT-/BSK-/CMD-)
package domain

import (
	"crypto/rand"
	"fmt"
)

const (
	length = 6
	charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	productPrefix = "PDT"
	cartPrefix = "BSK"
	orderPrefix = "CMD"
)

func generateSuffix(length int) (string, error) {
	suffix:= make([]byte, length)
	_, err := rand.Read(suffix)
	if err != nil {
		return "", fmt.Errorf("generateSuffix %d: %w", length, err)
	}

	for i, b := range suffix {
		suffix[i] = charset[int(b) % len(charset)]
	}
	return string(suffix), nil
}

func newID(prefix string) (string, error) {
	suffix, err := generateSuffix(length)
	if err != nil {
		return "", fmt.Errorf("newID %s: %w", prefix, err)
	}
	return prefix + "-" + suffix, nil
}

func NewProductID() (string, error) {
    return newID(productPrefix)
}

func NewCartID() (string, error) {
    return newID(cartPrefix)
}

func NewOrderID() (string, error) {
    return newID(orderPrefix)
}
