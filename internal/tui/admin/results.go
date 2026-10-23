package admin

import (
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func renderOrders(orders []apiclient.Order, width int) string {
	var s strings.Builder
	if len(orders) == 0 {
		s.WriteString(ui.InfoStyle.Render("Aucune commande."))
		s.WriteString("\n")
		return s.String()
	}

	inner := ui.CardInner(width)
	for _, o := range orders {
		left := ui.NameStyle.Render(o.ID) + ui.HelpStyle.Render("  "+ui.FormatDate(o.CreatedAt))

		body := []string{ui.HelpStyle.Render(o.UserEmail)}
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

var userStatusLabels = map[string]string{
	"pending":  "non confirmé",
	"active":   "actif",
	"disabled": "désactivé",
}

func userStatusLabel(s string) string {
	if l, ok := userStatusLabels[s]; ok {
		return l
	}
	return s
}

func roleLabel(r string) string {
	if r == "admin" {
		return "admin"
	}
	return "client"
}

func renderUsers(users []apiclient.AdminUser, width int) string {
	var s strings.Builder
	for _, user := range users {
		badge := ui.Badge(userStatusLabel(user.Status), ui.Green)
		if user.Status != "active" {
			badge = ui.Badge(userStatusLabel(user.Status), ui.Red)
		}
		left := ui.NameStyle.Render(user.Email)
		body := ui.HelpStyle.Render(fmt.Sprintf("#%d · %s", user.ID, roleLabel(user.Role)))
		s.WriteString(ui.Card(width, left, badge, body))
		s.WriteString("\n")
	}
	return s.String()
}
