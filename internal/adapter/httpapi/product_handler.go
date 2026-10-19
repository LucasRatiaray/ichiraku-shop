package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

type ProductHandler struct {
	products *usecase.ProductService
}

func NewProductHandler(products *usecase.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PriceHT     int64  `json:"price_ht"`
}

type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PriceHT     int64     `json:"price_ht"`
	PriceTTC    int64     `json:"price_ttc"`
	CreatedAt   time.Time `json:"created_at"`
}

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:          p.BusinessID,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		PriceHT:     p.PriceHT,
		PriceTTC:    p.PriceTTC(),
		CreatedAt:   p.CreatedAt,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	p, err := h.products.Create(r.Context(), req.Name, req.Description, req.Category, req.PriceHT)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toProductResponse(p))
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "name, category and positive price_ht are required")
	default:
		log.Printf("create product: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func int64Param(r *http.Request, name string) (int64, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return 0, nil
	}
	return strconv.ParseInt(v, 10, 64)
}

func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	f := domain.ProductFilter{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
	}

	params := []struct {
		name string
		dst  *int64
	}{
		{"min_price", &f.MinPriceHT},
		{"max_price", &f.MaxPriceHT},
		{"min_ttc", &f.MinPriceTTC},
		{"max_ttc", &f.MaxPriceTTC},
	}
	for _, p := range params {
		n, err := int64Param(r, p.name)
		if err != nil {
			writeError(w, http.StatusBadRequest, p.name+" must be an integer (cents)")
			return
		}
		*p.dst = n
	}

	products, err := h.products.Search(r.Context(), f)
	switch {
	case err == nil:
		resp := make([]productResponse, 0, len(products))
		for _, p := range products {
			resp = append(resp, toProductResponse(p))
		}
		writeJSON(w, http.StatusOK, resp)
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid price range")
	default:
		log.Printf("search products: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
