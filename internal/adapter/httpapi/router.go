package httpapi

import "net/http"

func NewRouter(auth *AuthHandler, products *ProductHandler, baskets *BasketHandler, orders *OrderHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /auth/register", auth.Register)

	mux.HandleFunc("POST /auth/confirm", auth.Confirm)

	mux.HandleFunc("POST /auth/login", auth.Login)

	mux.HandleFunc("GET /auth/me", auth.RequireAuth(auth.Me))

	mux.HandleFunc("POST /auth/password/forgot", auth.ForgotPassword)

	mux.HandleFunc("POST /auth/password/reset", auth.ResetPassword)

	mux.HandleFunc("POST /auth/logout", auth.RequireAuth(auth.Logout))

	mux.HandleFunc("POST /admin/products", auth.RequireAdmin(products.Create))

	mux.HandleFunc("GET /products", products.Search)

	mux.HandleFunc("GET /basket", auth.RequireAuth(baskets.Get))
	mux.HandleFunc("POST /basket/items", auth.RequireAuth(baskets.AddItem))
	mux.HandleFunc("PUT /basket/items/{productID}", auth.RequireAuth(baskets.SetQuantity))
	mux.HandleFunc("DELETE /basket/items/{productID}", auth.RequireAuth(baskets.RemoveItem))

	mux.HandleFunc("POST /orders", auth.RequireAuth(orders.Checkout))

	return mux
}
