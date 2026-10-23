package domain

import (
	"errors"
	"testing"
)

func TestChangeStatus(t *testing.T) {
	tests := []struct {
		name       string
		from       OrderStatus
		to         OrderStatus
		reason     string
		wantErr    error
		wantStatus OrderStatus
		wantReason string
	}{
		{"en attente vers payé", OrderPending, OrderPaid, "", nil, OrderPaid, ""},
		{"payé vers en livraison", OrderPaid, OrderShipping, "", nil, OrderShipping, ""},
		{"en livraison vers livré", OrderShipping, OrderDelivered, "", nil, OrderDelivered, ""},
		{"retour en arrière interdit", OrderShipping, OrderPaid, "", ErrInvalidTransition, OrderShipping, ""},
		{"état final livré", OrderDelivered, OrderCancelled, "raison", ErrInvalidTransition, OrderDelivered, ""},
		{"annulation sans raison", OrderPaid, OrderCancelled, "", ErrCancelReasonRequired, OrderPaid, ""},
		{"annulation avec raison à nettoyer", OrderPaid, OrderCancelled, "  Rupture  ", nil, OrderCancelled, "Rupture"},
		{"statut inconnu", OrderPaid, OrderStatus("lost"), "", ErrInvalidStatus, OrderPaid, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Order{Status: tt.from}
			err := o.ChangeStatus(tt.to, tt.reason)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("erreur = %v, attendu %v", err, tt.wantErr)
			}
			if o.Status != tt.wantStatus {
				t.Errorf("statut = %q, attendu %q", o.Status, tt.wantStatus)
			}
			if o.CancelReason != tt.wantReason {
				t.Errorf("raison = %q, attendu %q", o.CancelReason, tt.wantReason)
			}
		})
	}
}

func TestNewOrderFromBasket(t *testing.T) {
	t.Run("panier vide", func(t *testing.T) {
		_, err := NewOrderFromBasket(Basket{UserID: 1}, OrderPaid)
		if !errors.Is(err, ErrEmptyBasket) {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyBasket)
		}
	})

	t.Run("panier de deux lignes", func(t *testing.T) {
		basket := Basket{
			UserID: 7,
			Items: []BasketItem{
				{Product: Product{ID: 1, BusinessID: "PDT-AAAAAA", Name: "Ramen", PriceHT: 1000}, Quantity: 2},
				{Product: Product{ID: 2, BusinessID: "PDT-BBBBBB", Name: "Gyoza", PriceHT: 700}, Quantity: 1},
			},
		}

		order, err := NewOrderFromBasket(basket, OrderPaid)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if want := int64(2*1200 + 840); order.TotalTTC != want {
			t.Errorf("total = %d, attendu %d", order.TotalTTC, want)
		}
		if order.UserID != 7 || order.Status != OrderPaid {
			t.Errorf("utilisateur %d, statut %q: attendu 7 et %q", order.UserID, order.Status, OrderPaid)
		}
		if len(order.Items) != 2 {
			t.Fatalf("%d lignes, attendu 2", len(order.Items))
		}

		want := []OrderItem{
			{ProductID: 1, ProductBusinessID: "PDT-AAAAAA", Name: "Ramen", UnitPriceTTC: 1200, Quantity: 2},
			{ProductID: 2, ProductBusinessID: "PDT-BBBBBB", Name: "Gyoza", UnitPriceTTC: 840, Quantity: 1},
		}
		for i, w := range want {
			if order.Items[i] != w {
				t.Errorf("ligne %d = %+v, attendu %+v", i, order.Items[i], w)
			}
		}
	})
}
