package client

import (
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
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
