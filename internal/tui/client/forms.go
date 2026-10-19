package client

import (
	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

func newMenuForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("choice").
				Options(
					huh.NewOption("Se connecter", "login"),
					huh.NewOption("Créer un compte", "register"),
					huh.NewOption("Mot de passe oublié", "forgot"),
					huh.NewOption("Quitter", "quit"),
				),
		),
	)
}

func newHomeForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("choice").
				Title("Bienvenue").
				Options(
					huh.NewOption("Catalogue", "catalog"),
					huh.NewOption("Mon panier", "cart"),
					huh.NewOption("Se déconnecter", "logout"),
					huh.NewOption("Quitter", "quit"),
				),
		),
	)
}

func newLoginForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email"),
			huh.NewInput().Key("password").Title("Mot de passe").EchoMode(huh.EchoModePassword),
		),
	)
}

func newRegisterForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email"),
			huh.NewInput().Key("password").Title("Mot de passe (8 caractères min.)").EchoMode(huh.EchoModePassword),
		),
	)
}

func newConfirmForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("code").Title("Code reçu par email"),
		),
	)
}

func newForgotForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email du compte"),
		),
	)
}

func newResetForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("code").Title("Code reçu par email"),
			huh.NewInput().Key("password").Title("Nouveau mot de passe (8 caractères min.)").EchoMode(huh.EchoModePassword),
		),
	)
}

func newSearchForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("q").Title("Nom ou description").Placeholder("vide = tout"),
			huh.NewInput().Key("category").Title("Catégorie"),
			huh.NewInput().Key("min_ht").Title("Prix HT min (€)").Validate(validatePrice),
			huh.NewInput().Key("max_ht").Title("Prix HT max (€)").Validate(validatePrice),
			huh.NewInput().Key("min_ttc").Title("Prix TTC min (€)").Validate(validatePrice),
			huh.NewInput().Key("max_ttc").Title("Prix TTC max (€)").Validate(validatePrice),
		),
	)
}

func filterFromSearchForm(form *huh.Form) domain.ProductFilter {
	minHT, _ := parseEuros(form.GetString("min_ht"))
	maxHT, _ := parseEuros(form.GetString("max_ht"))
	minTTC, _ := parseEuros(form.GetString("min_ttc"))
	maxTTC, _ := parseEuros(form.GetString("max_ttc"))

	return domain.ProductFilter{
		Query:       form.GetString("q"),
		Category:    form.GetString("category"),
		MinPriceHT:  minHT,
		MaxPriceHT:  maxHT,
		MinPriceTTC: minTTC,
		MaxPriceTTC: maxTTC,
	}
}
