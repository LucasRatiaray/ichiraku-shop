package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

type OrderHandler struct {
	orders *usecase.OrderService
}

func NewOrderHandler(orders *usecase.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type checkoutRequest struct {
	CardNumber string `json:"card_number"`
	Expiry     string `json:"expiry"`
	CVC        string `json:"cvc"`
}

type orderItemResponse struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	UnitPriceTTC int64  `json:"unit_price_ttc"`
	Quantity     int    `json:"quantity"`
	TotalTTC     int64  `json:"total_ttc"`
}

type orderResponse struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	CancelReason string              `json:"cancel_reason,omitempty"`
	TotalTTC     int64               `json:"total_ttc"`
	Items        []orderItemResponse `json:"items"`
	CreatedAt    time.Time           `json:"created_at"`
	UserEmail    string              `json:"user_email"`
}

func toOrderResponse(o domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemResponse{
			ProductID:    it.ProductBusinessID,
			Name:         it.Name,
			UnitPriceTTC: it.UnitPriceTTC,
			Quantity:     it.Quantity,
			TotalTTC:     it.TotalTTC(),
		})
	}
	return orderResponse{
		ID:           o.BusinessID,
		Status:       string(o.Status),
		CancelReason: o.CancelReason,
		TotalTTC:     o.TotalTTC,
		Items:        items,
		CreatedAt:    o.CreatedAt,
		UserEmail:    o.UserEmail,
	}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	o, err := h.orders.Pay(r.Context(), user.ID, req.CardNumber, req.Expiry, req.CVC)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toOrderResponse(o))
	case errors.Is(err, domain.ErrInvalidCardNumber),
		errors.Is(err, domain.ErrInvalidExpiry),
		errors.Is(err, domain.ErrCardExpired),
		errors.Is(err, domain.ErrInvalidCVC),
		errors.Is(err, domain.ErrEmptyBasket):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("checkout: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	orders, err := h.orders.ListMine(r.Context(), user.ID)
	if err != nil {
		log.Printf("list orders: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, toOrderResponse(o))
	}
	writeJSON(w, http.StatusOK, resp)
}

type changeStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func writeOrders(w http.ResponseWriter, orders []domain.Order) {
	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, toOrderResponse(o))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orders.ListAll(r.Context())
	if err != nil {
		log.Printf("list all orders: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeOrders(w, orders)
}

func (h *OrderHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req changeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	o, err := h.orders.ChangeStatus(r.Context(), r.PathValue("orderID"), domain.OrderStatus(req.Status), req.Reason)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toOrderResponse(o))
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "order not found")
	case errors.Is(err, domain.ErrInvalidStatus), errors.Is(err, domain.ErrCancelReasonRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "order was modified meanwhile, reload and retry")
	default:
		log.Printf("change order status: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

type adminOrderLine struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type adminCreateOrderRequest struct {
	UserEmail string           `json:"user_email"`
	Items     []adminOrderLine `json:"items"`
}

func (h *OrderHandler) CreateForUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req adminCreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	lines := make([]usecase.OrderLine, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, usecase.OrderLine{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	o, err := h.orders.CreateForUser(r.Context(), req.UserEmail, lines)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toOrderResponse(o))
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "at least one item, quantities between 1 and 99")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "user or product not found")
	default:
		log.Printf("admin create order: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
