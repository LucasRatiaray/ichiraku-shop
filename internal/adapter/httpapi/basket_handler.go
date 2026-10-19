package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

type BasketHandler struct {
	baskets *usecase.BasketService
}

func NewBasketHandler(baskets *usecase.BasketService) *BasketHandler {
	return &BasketHandler{baskets: baskets}
}

type addItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type setQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type basketItemResponse struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	UnitPriceTTC int64  `json:"unit_price_ttc"`
	Quantity     int    `json:"quantity"`
	TotalTTC     int64  `json:"total_ttc"`
}

type basketResponse struct {
	ID           string               `json:"id"`
	Items        []basketItemResponse `json:"items"`
	ShippingCost int64                `json:"shipping_cost"`
	TotalTTC     int64                `json:"total_ttc"`
}

func toBasketResponse(b domain.Basket) basketResponse {
	items := make([]basketItemResponse, 0, len(b.Items))
	for _, it := range b.Items {
		items = append(items, basketItemResponse{
			ProductID:    it.Product.BusinessID,
			Name:         it.Product.Name,
			UnitPriceTTC: it.Product.PriceTTC(),
			Quantity:     it.Quantity,
			TotalTTC:     it.TotalTTC(),
		})
	}
	return basketResponse{
		ID:           b.BusinessID,
		Items:        items,
		ShippingCost: domain.ShippingCost,
		TotalTTC:     b.TotalTTC(),
	}
}

func writeBasketResult(w http.ResponseWriter, b domain.Basket, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toBasketResponse(b))
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "quantity must be between 1 and 99")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "product not found")
	default:
		log.Printf("basket: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *BasketHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	b, err := h.baskets.Get(r.Context(), user.ID)
	writeBasketResult(w, b, err)
}

func (h *BasketHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, err := h.baskets.AddItem(r.Context(), user.ID, req.ProductID, req.Quantity)
	writeBasketResult(w, b, err)
}

func (h *BasketHandler) SetQuantity(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req setQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, err := h.baskets.SetQuantity(r.Context(), user.ID, r.PathValue("productID"), req.Quantity)
	writeBasketResult(w, b, err)
}

func (h *BasketHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	b, err := h.baskets.RemoveItem(r.Context(), user.ID, r.PathValue("productID"))
	writeBasketResult(w, b, err)
}