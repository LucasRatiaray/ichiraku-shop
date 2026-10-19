package client

import (
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func renderProducts(products []apiclient.Product) string {
	if len(products) == 0 {
		return infoStyle.Render("Aucun produit trouvé.") + "\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d produit(s)\n\n", len(products))
	for _, p := range products {
		fmt.Fprintf(&b, "%s  %s  [%s]\n", helpStyle.Render(p.ID), nameStyle.Render(p.Name), p.Category)
		fmt.Fprintf(&b, "    %s HT · %s TTC\n", formatEuros(p.PriceHT), formatEuros(p.PriceTTC))
		if p.Description != "" {
			fmt.Fprintf(&b, "    %s\n", helpStyle.Render(p.Description))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderBasket(b apiclient.Basket) string {
	var s strings.Builder
	fmt.Fprintf(&s, "Panier %s\n\n", ui.HelpStyle.Render(b.ID))

	if len(b.Items) == 0 {
		s.WriteString(ui.InfoStyle.Render("Ton panier est vide."))
		s.WriteString("\n")
		return s.String()
	}

	for _, it := range b.Items {
		fmt.Fprintf(&s, "%s  ×%d  %s\n", ui.NameStyle.Render(it.Name), it.Quantity, ui.FormatEuros(it.TotalTTC))
		fmt.Fprintf(&s, "    %s · %s TTC l'unité\n", ui.HelpStyle.Render(it.ProductID), ui.FormatEuros(it.UnitPriceTTC))
	}
	fmt.Fprintf(&s, "\nLivraison : gratuite\n")
	fmt.Fprintf(&s, "Total TTC : %s\n", ui.NameStyle.Render(ui.FormatEuros(b.TotalTTC)))
	return s.String()
}

func renderOrders(orders []apiclient.Order) string {
	var s strings.Builder
	if len(orders) == 0 {
		s.WriteString(ui.InfoStyle.Render("Aucune commande pour l'instant."))
		s.WriteString("\n")
		return s.String()
	}

	for _, o := range orders {
		status := ui.InfoStyle.Render(ui.OrderStatusLabel(o.Status))
		if o.Status == "cancelled" {
			status = ui.ErrStyle.Render(ui.OrderStatusLabel(o.Status))
		}
		fmt.Fprintf(&s, "%s  %s  %s\n", ui.NameStyle.Render(o.ID), ui.FormatDate(o.CreatedAt), status)
		if o.CancelReason != "" {
			fmt.Fprintf(&s, "    Raison : %s\n", o.CancelReason)
		}
		for _, it := range o.Items {
			fmt.Fprintf(&s, "    %d × %s  %s\n", it.Quantity, it.Name, ui.FormatEuros(it.TotalTTC))
		}
		fmt.Fprintf(&s, "    Total : %s TTC\n\n", ui.FormatEuros(o.TotalTTC))
	}
	return s.String()
}
