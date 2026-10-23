package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestTabsFitWidth(t *testing.T) {
	short := []Tab{
		{Key: "s", Label: "boutique", Active: true},
		{Key: "c", Label: "panier", Detail: "0,00 € [0]"},
		{Key: "o", Label: "commandes"},
		{Key: "u", Label: "compte"},
		{Key: "q", Label: "quitter"},
	}
	long := []Tab{
		{Key: "s", Label: "boutique"},
		{Key: "c", Label: "panier", Detail: "1250,00 € [12]", Active: true},
		{Key: "o", Label: "commandes"},
		{Key: "u", Label: "compte"},
		{Key: "q", Label: "quitter"},
	}

	tests := []struct {
		name  string
		tabs  []Tab
		width int
	}{
		{"panier vide à 76 colonnes", short, 76},
		{"panier vide à 56 colonnes", short, 56},
		{"gros panier à 76 colonnes", long, 76},
		{"gros panier à 56 colonnes", long, 56},
		{"terminal très étroit", long, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := Tabs(tt.width, tt.tabs)
			lines := strings.Split(bar, "\n")
			if len(lines) != 3 {
				t.Errorf("%d lignes, attendu 3", len(lines))
			}
			for i, l := range lines {
				if w := lipgloss.Width(l); w > tt.width {
					t.Errorf("ligne %d de %d colonnes, maximum %d", i, w, tt.width)
				}
			}
		})
	}

	t.Run("libellés visibles quand la place suffit", func(t *testing.T) {
		if bar := Tabs(76, short); !strings.Contains(bar, "boutique") {
			t.Error("les libellés devraient être affichés à 76 colonnes")
		}
	})

	t.Run("touches seules quand la place manque", func(t *testing.T) {
		bar := Tabs(56, long)
		if strings.Contains(bar, "boutique") {
			t.Error("les libellés ne devraient plus être affichés à 56 colonnes")
		}
		if !strings.Contains(bar, "1250,00 € [12]") {
			t.Error("la valeur du panier doit rester visible en mode compact")
		}
	})
}
