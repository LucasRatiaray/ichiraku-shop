package ui

import (
	"image/color"
	"time"
	_ "time/tzdata"
)

var statusLabels = map[string]string{
	"pending":   "En attente",
	"paid":      "Payée",
	"shipping":  "En cours de livraison",
	"delivered": "Livrée",
	"cancelled": "Annulée",
}

func OrderStatusLabel(status string) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return status
}

var statusColors = map[string]color.Color{
	"pending":   Yellow,
	"paid":      Blue,
	"shipping":  Accent,
	"delivered": Green,
	"cancelled": Red,
}

func OrderStatusBadge(status string) string {
	bg, ok := statusColors[status]
	if !ok {
		bg = Muted
	}
	return Badge(OrderStatusLabel(status), bg)
}

var paris = loadParis()

func loadParis() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return time.UTC
	}
	return loc
}

func FormatDate(t time.Time) string {
	return t.In(paris).Format("02/01/2006 15:04")
}
