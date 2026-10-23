package usecase

import (
	"errors"
	"strconv"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

var errMissingVersion = errors.New("token without version claim")

type tokenClaims struct {
	Role    string `json:"role"`
	Version *int   `json:"ver"`
	jwt.RegisteredClaims
}

func signToken(secret []byte, user domain.User, now time.Time) (string, error) {
	claims := tokenClaims{
		Role:    string(user.Role),
		Version: &user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func parseToken(secret []byte, token string) (userID int64, version int, err error) {
	var claims tokenClaims
	_, err = jwt.ParseWithClaims(token, &claims,
		func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, 0, err
	}
	if claims.Version == nil {
		return 0, 0, errMissingVersion
	}
	userID, err = strconv.ParseInt(claims.Subject, 10, 64)
	return userID, *claims.Version, err
}
