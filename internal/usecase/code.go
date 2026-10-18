package usecase

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}
