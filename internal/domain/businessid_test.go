package domain

import (
	"regexp"
	"strings"
	"testing"
)

func TestNewBusinessID(t *testing.T) {
	t.Run("format PDT-XXXXXX", func(t *testing.T) {
		id, err := NewBusinessID(PrefixProduct)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if len(id) != 10 {
			t.Errorf("longueur = %d, attendu 10", len(id))
		}
		if !strings.HasPrefix(id, "PDT-") {
			t.Errorf("préfixe de %q, attendu PDT-", id)
		}
		if !regexp.MustCompile(`^PDT-[0-9A-Z]{6}$`).MatchString(id) {
			t.Errorf("%q ne correspond pas à ^PDT-[0-9A-Z]{6}$", id)
		}
	})

	t.Run("1000 générations dans l'alphabet", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			id, err := NewBusinessID(PrefixOrder)
			if err != nil {
				t.Fatalf("erreur inattendue: %v", err)
			}
			for _, c := range strings.TrimPrefix(id, "CMD-") {
				if !strings.ContainsRune(businessIDAlphabet, c) {
					t.Fatalf("caractère %q hors alphabet dans %q", c, id)
				}
			}
		}
	})

	t.Run("préfixes panier et commande", func(t *testing.T) {
		tests := []struct {
			prefix string
			re     string
		}{
			{PrefixBasket, `^BSK-[0-9A-Z]{6}$`},
			{PrefixOrder, `^CMD-[0-9A-Z]{6}$`},
		}
		for _, tt := range tests {
			id, err := NewBusinessID(tt.prefix)
			if err != nil {
				t.Fatalf("erreur inattendue: %v", err)
			}
			if !regexp.MustCompile(tt.re).MatchString(id) {
				t.Errorf("%q ne correspond pas à %s", id, tt.re)
			}
		}
	})
}
