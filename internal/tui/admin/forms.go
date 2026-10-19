package admin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
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

func newHomeForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("choice").Title("Administration").Options(
				huh.NewOption("Créer un produit", "create_product"),
				huh.NewOption("Commandes", "orders"),
				huh.NewOption("Créer une commande pour un client", "create_order"),
				huh.NewOption("Se déconnecter", "logout"),
				huh.NewOption("Quitter", "quit"),
			),
		),
	)
}

func validateQuantity(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 99 {
		return errors.New("quantité entre 1 et 99")
	}
	return nil
}

func newChangeStatusForm(orders []apiclient.Order) *huh.Form {
	orderOpts := make([]huh.Option[string], 0, len(orders))
	for _, o := range orders {
		label := fmt.Sprintf("%s — %s — %s", o.ID, o.UserEmail, ui.OrderStatusLabel(o.Status))
		orderOpts = append(orderOpts, huh.NewOption(label, o.ID))
	}

	statuses := []string{"pending", "paid", "shipping", "delivered", "cancelled"}
	statusOpts := make([]huh.Option[string], 0, len(statuses))
	for _, s := range statuses {
		statusOpts = append(statusOpts, huh.NewOption(ui.OrderStatusLabel(s), s))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("order").Title("Commande").Options(orderOpts...),
			huh.NewSelect[string]().Key("status").Title("Nouveau statut").Options(statusOpts...),
			huh.NewInput().Key("reason").Title("Raison (obligatoire si annulation)"),
		),
	)
}

func newCreateOrderForm(products []apiclient.Product) *huh.Form {
	opts := make([]huh.Option[string], 0, len(products))
	for _, p := range products {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s — %s TTC", p.Name, ui.FormatEuros(p.PriceTTC)), p.ID))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email du client").Validate(required),
			huh.NewMultiSelect[string]().Key("products").Title("Produits").Options(opts...).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("choisis au moins un produit")
					}
					return nil
				}),
			huh.NewInput().Key("quantity").Title("Quantité par produit").Placeholder("1").Validate(validateQuantity),
		),
	)
}