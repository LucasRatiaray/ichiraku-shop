package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

func TestSignAndParseToken(t *testing.T) {
	secret := []byte("secret-de-test-0123456789-abcdefghijklmnop")
	other := []byte("un-autre-secret-de-plus-de-32-caracteres!!")
	user := domain.User{ID: 42, Role: domain.RoleAdmin, TokenVersion: 3}
	now := time.Now()

	t.Run("aller-retour avec le même secret", func(t *testing.T) {
		token, err := signToken(secret, user, now)
		if err != nil {
			t.Fatalf("signToken: %v", err)
		}
		id, version, err := parseToken(secret, token)
		if err != nil {
			t.Fatalf("parseToken: %v", err)
		}
		if id != 42 || version != 3 {
			t.Errorf("sub = %d, ver = %d, attendu 42 et 3", id, version)
		}
	})

	t.Run("rôle présent dans les claims", func(t *testing.T) {
		token, err := signToken(secret, user, now)
		if err != nil {
			t.Fatalf("signToken: %v", err)
		}
		var claims tokenClaims
		if _, _, err := jwt.NewParser().ParseUnverified(token, &claims); err != nil {
			t.Fatalf("lecture des claims: %v", err)
		}
		if claims.Role != "admin" {
			t.Errorf("rôle = %q, attendu admin", claims.Role)
		}
	})

	t.Run("autre secret", func(t *testing.T) {
		token, err := signToken(other, user, now)
		if err != nil {
			t.Fatalf("signToken: %v", err)
		}
		if _, _, err := parseToken(secret, token); err == nil {
			t.Error("un token signé avec un autre secret a été accepté")
		}
	})

	t.Run("token expiré", func(t *testing.T) {
		token, err := signToken(secret, user, now.Add(-2*tokenTTL))
		if err != nil {
			t.Fatalf("signToken: %v", err)
		}
		if _, _, err := parseToken(secret, token); err == nil {
			t.Error("un token expiré a été accepté")
		}
	})

	t.Run("signature modifiée d'un caractère", func(t *testing.T) {
		token, err := signToken(secret, user, now)
		if err != nil {
			t.Fatalf("signToken: %v", err)
		}
		parts := strings.Split(token, ".")
		sig := []byte(parts[2])
		i := len(sig) / 2
		if sig[i] == 'A' {
			sig[i] = 'B'
		} else {
			sig[i] = 'A'
		}
		parts[2] = string(sig)
		if _, _, err := parseToken(secret, strings.Join(parts, ".")); err == nil {
			t.Error("un token modifié a été accepté")
		}
	})
}
