// (génération PDT-/BSK-/CMD-)
package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const businessIDAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	PrefixProduct = "PDT"
	PrefixBasket  = "BSK"
	PrefixOrder   = "CMD"
)

func NewBusinessID(prefix string) (string, error) {
	b := make([]byte, 6)
	max := big.NewInt(int64(len(businessIDAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("business id: %w", err)
		}
		b[i] = businessIDAlphabet[n.Int64()]
	}
	return prefix + "-" + string(b), nil
}
