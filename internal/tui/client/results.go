package client

import (
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func renderProducts(products []apiclient.Product, width int) string {
	if len(products) == 0 {
		return ui.InfoStyle.Render("Aucun produit trouvé.") + "\n"
	}

	var b strings.Builder
	b.WriteString(ui.HelpStyle.Render(fmt.Sprintf("%d produit(s)", len(products))))
	b.WriteString("\n")
	for _, p := range products {
		body := ui.HelpStyle.Render(fmt.Sprintf("%s · %s · %s HT", p.ID, p.Category, ui.FormatEuros(p.PriceHT)))
		if p.Description != "" {
			body += "\n" + p.Description
		}
		b.WriteString(ui.Card(width, ui.NameStyle.Render(p.Name), ui.PriceStyle.Render(ui.FormatEuros(p.PriceTTC)+" TTC"), body))
		b.WriteString("\n")
	}
	return b.String()
}

func renderBasket(b apiclient.Basket, width int) string {
	var s strings.Builder
	s.WriteString(ui.HelpStyle.Render("Panier " + b.ID))
	s.WriteString("\n")

	if len(b.Items) == 0 {
		s.WriteString(ui.InfoStyle.Render("Ton panier est vide."))
		s.WriteString("\n")
		return s.String()
	}

	for _, it := range b.Items {
		left := ui.NameStyle.Render(it.Name) + ui.HelpStyle.Render(fmt.Sprintf("  ×%d", it.Quantity))
		body := ui.HelpStyle.Render(fmt.Sprintf("%s · %s TTC l'unité", it.ProductID, ui.FormatEuros(it.UnitPriceTTC)))
		s.WriteString(ui.Card(width, left, ui.PriceStyle.Render(ui.FormatEuros(it.TotalTTC)), body))
		s.WriteString("\n")
	}
	s.WriteString(ui.Row(width, ui.HelpStyle.Render("Livraison"), "gratuite"))
	s.WriteString("\n")
	s.WriteString(ui.Row(width, ui.NameStyle.Render("Total TTC"), ui.PriceStyle.Render(ui.FormatEuros(b.TotalTTC))))
	s.WriteString("\n")
	return s.String()
}

func renderOrders(orders []apiclient.Order, width int) string {
	var s strings.Builder
	if len(orders) == 0 {
		s.WriteString(ui.InfoStyle.Render("Aucune commande pour l'instant."))
		s.WriteString("\n")
		return s.String()
	}

	inner := ui.CardInner(width)
	for _, o := range orders {
		left := ui.NameStyle.Render(o.ID) + ui.HelpStyle.Render("  "+ui.FormatDate(o.CreatedAt))

		var body []string
		if o.CancelReason != "" {
			body = append(body, ui.ErrStyle.Render("Raison : "+o.CancelReason))
		}
		for _, it := range o.Items {
			body = append(body, ui.Row(inner, fmt.Sprintf("%d × %s", it.Quantity, it.Name), ui.FormatEuros(it.TotalTTC)))
		}
		body = append(body, ui.Row(inner, ui.HelpStyle.Render("Total TTC"), ui.PriceStyle.Render(ui.FormatEuros(o.TotalTTC))))

		s.WriteString(ui.Card(width, left, ui.OrderStatusBadge(o.Status), strings.Join(body, "\n")))
		s.WriteString("\n")
	}
	return s.String()
}
