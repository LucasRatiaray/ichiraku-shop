package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

type userCtxKey struct{}

func (h *AuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}

		user, err := h.auth.Authenticate(r.Context(), token)
		switch {
		case errors.Is(err, domain.ErrUnauthorized):
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		case err != nil:
			log.Printf("auth middleware: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey{}, user)

		next(w, r.WithContext(ctx))
	}
}

func userFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(userCtxKey{}).(domain.User)
	return user, ok
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	return token, ok && token != ""
}
