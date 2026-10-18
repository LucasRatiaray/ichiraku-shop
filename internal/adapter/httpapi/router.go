package httpapi

import "net/http"

func NewRouter(auth *AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /auth/register", auth.Register)

	mux.HandleFunc("POST /auth/confirm", auth.Confirm)

	mux.HandleFunc("POST /auth/login", auth.Login)

	return mux
}
