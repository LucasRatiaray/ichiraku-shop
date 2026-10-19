package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b) // remplit le slice b avec des octets aléatoires
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token)) // empreinte du token
	return hex.EncodeToString(sum[:])
}
