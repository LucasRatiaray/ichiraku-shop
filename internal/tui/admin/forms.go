package admin

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("champ obligatoire")
	}
	return nil
}

func requiredPrice(s string) error {
	cents, err := ui.ParseEuros(s)
	if err != nil {
		return err
	}
	if cents <= 0 {
		return errors.New("prix obligatoire et supérieur à 0")
	}
	return nil
}

func newLoginForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email admin"),
			huh.NewInput().Key("password").Title("Mot de passe").EchoMode(huh.EchoModePassword),
		),
	)
}

func newHomeForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("choice").Title("Administration").Options(
				huh.NewOption("Créer un produit", "create_product"),
				huh.NewOption("Se déconnecter", "logout"),
				huh.NewOption("Quitter", "quit"),
			),
		),
	)
}

func newCreateProductForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("name").Title("Nom").Validate(required),
			huh.NewInput().Key("description").Title("Description (optionnelle)"),
			huh.NewInput().Key("category").Title("Catégorie").Validate(required),
			huh.NewInput().Key("price").Title("Prix HT (€)").Placeholder("12,50").Validate(requiredPrice),
		),
	)
}