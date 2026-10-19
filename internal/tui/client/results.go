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
