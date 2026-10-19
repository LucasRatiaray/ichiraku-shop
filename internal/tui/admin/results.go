package admin

import (
	"fmt"
	"strings"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func renderOrders(orders []apiclient.Order) string {
	var s strings.Builder
	if len(orders) == 0 {
		s.WriteString(ui.InfoStyle.Render("Aucune commande."))
		s.WriteString("\n")
		return s.String()
	}

	for _, o := range orders {
		status := ui.InfoStyle.Render(ui.OrderStatusLabel(o.Status))
		if o.Status == "cancelled" {
			status = ui.ErrStyle.Render(ui.OrderStatusLabel(o.Status))
		}
		fmt.Fprintf(&s, "%s  %s  %s  %s\n", ui.NameStyle.Render(o.ID), o.UserEmail, ui.FormatDate(o.CreatedAt), status)
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

func renderUsers(users []apiclient.AdminUser) string {
	var s strings.Builder
	for _, user := range users {
		status := ui.InfoStyle.Render(userStatusLabel(user.Status))
		if user.Status != "active" {
			status = ui.ErrStyle.Render(userStatusLabel(user.Status))
		}
		fmt.Fprintf(&s, "#%-4d %s  %s  %s\n", user.ID, ui.NameStyle.Render(user.Email), roleLabel(user.Role), status)
	}
	return s.String()
}
