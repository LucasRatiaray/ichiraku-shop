package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

func newMenuForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("choice").
				Options(
					huh.NewOption("Se connecter", "login"),
					huh.NewOption("Créer un compte", "register"),
					huh.NewOption("Renvoyer le code de confirmation", "resend"),
					huh.NewOption("Mot de passe oublié", "forgot"),
					huh.NewOption("Quitter", "quit"),
				),
		),
	)
}

func newLoginForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email"),
			huh.NewInput().Key("password").Title("Mot de passe").EchoMode(huh.EchoModePassword),
		),
	)
}

func newRegisterForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email"),
			huh.NewInput().Key("password").Title("Mot de passe (8 caractères min.)").EchoMode(huh.EchoModePassword),
		),
	)
}

func newConfirmForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("code").Title("Code reçu par email"),
		),
	)
}

func newForgotForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email du compte"),
		),
	)
}

func newResetForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("code").Title("Code reçu par email"),
			huh.NewInput().Key("password").Title("Nouveau mot de passe (8 caractères min.)").EchoMode(huh.EchoModePassword),
		),
	)
}

func newSearchForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("q").Title("Nom ou description").Placeholder("vide = tout"),
			huh.NewInput().Key("category").Title("Catégorie"),
			huh.NewInput().Key("min_ht").Title("Prix HT min (€)").Validate(ui.ValidatePrice),
			huh.NewInput().Key("max_ht").Title("Prix HT max (€)").Validate(ui.ValidatePrice),
			huh.NewInput().Key("min_ttc").Title("Prix TTC min (€)").Validate(ui.ValidatePrice),
			huh.NewInput().Key("max_ttc").Title("Prix TTC max (€)").Validate(ui.ValidatePrice),
		),
	)
}

func filterFromSearchForm(form *huh.Form) domain.ProductFilter {
	minHT, _ := ui.ParseEuros(form.GetString("min_ht"))
	maxHT, _ := ui.ParseEuros(form.GetString("max_ht"))
	minTTC, _ := ui.ParseEuros(form.GetString("min_ttc"))
	maxTTC, _ := ui.ParseEuros(form.GetString("max_ttc"))

	return domain.ProductFilter{
		Query:       form.GetString("q"),
		Category:    form.GetString("category"),
		MinPriceHT:  minHT,
		MaxPriceHT:  maxHT,
		MinPriceTTC: minTTC,
		MaxPriceTTC: maxTTC,
	}
}

func validateQuantity(min int) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil || n < min || n > 99 {
			return fmt.Errorf("quantité entre %d et 99", min)
		}
		return nil
	}
}

func newAddToBasketForm(products []apiclient.Product) *huh.Form {
	opts := make([]huh.Option[string], 0, len(products))
	for _, p := range products {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s — %s TTC", p.Name, ui.FormatEuros(p.PriceTTC)), p.ID))
	}
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("product").Title("Produit").Options(opts...),
			huh.NewInput().Key("quantity").Title("Quantité").Placeholder("1").Validate(validateQuantity(1)),
		),
	)
}

func newEditBasketForm(items []apiclient.BasketItem) *huh.Form {
	opts := make([]huh.Option[string], 0, len(items))
	for _, it := range items {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%d)", it.Name, it.Quantity), it.ProductID))
	}
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("product").Title("Ligne à modifier").Options(opts...),
			huh.NewInput().Key("quantity").Title("Nouvelle quantité (0 = supprimer)").Validate(validateQuantity(0)),
		),
	)
}

func newPaymentForm(total int64) *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("card_number").Title("Numéro de carte").Placeholder("4242 4242 4242 4242").Validate(required),
			huh.NewInput().Key("expiry").Title("Expiration").Placeholder("MM/AA").Validate(required),
			huh.NewInput().Key("cvc").Title("CVC").EchoMode(huh.EchoModePassword).Validate(required),
		).Title("Paiement — " + ui.FormatEuros(total) + " TTC"),
	)
}

func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("champ obligatoire")
	}
	return nil
}
