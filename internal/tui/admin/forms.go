package admin

import (
	"errors"
	"fmt"
	"net/mail"
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
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email admin"),
			huh.NewInput().Key("password").Title("Mot de passe").EchoMode(huh.EchoModePassword),
		),
	)
}

func newCreateProductForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("name").Title("Nom").Validate(required),
			huh.NewInput().Key("description").Title("Description (optionnelle)"),
			huh.NewInput().Key("category").Title("Catégorie").Validate(required),
			huh.NewInput().Key("price").Title("Prix HT (€)").Placeholder("12,50").Validate(requiredPrice),
		),
	)
}

func newHomeForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("choice").Title("Administration").Options(
				huh.NewOption("Créer un produit", "create_product"),
				huh.NewOption("Commandes", "orders"),
				huh.NewOption("Créer une commande pour un client", "create_order"),
				huh.NewOption("Utilisateurs", "users"),
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

	return ui.NewForm(
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

	return ui.NewForm(
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

func validatePassword(s string) error {
	if len(s) < 8 || len(s) > 72 {
		return errors.New("8 à 72 caractères")
	}
	return nil
}

func newCreateUserForm() *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email").Validate(required),
			huh.NewInput().Key("password").Title("Mot de passe initial").EchoMode(huh.EchoModePassword).Validate(validatePassword),
			huh.NewSelect[string]().Key("role").Title("Rôle").Options(
				huh.NewOption("Client", "user"),
				huh.NewOption("Admin", "admin"),
			),
		),
	)
}

func optionalEmail(s string) error {
	email := strings.ToLower(strings.TrimSpace(s))
	if email == "" {
		return nil
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("email invalide")
	}
	return nil
}

func optionalPassword(s string) error {
	if s == "" {
		return nil
	}
	return validatePassword(s)
}

func userOptions(users []apiclient.AdminUser) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(users))
	for _, user := range users {
		label := fmt.Sprintf("#%d %s (%s, %s)", user.ID, user.Email, roleLabel(user.Role), userStatusLabel(user.Status))
		opts = append(opts, huh.NewOption(label, strconv.FormatInt(user.ID, 10)))
	}
	return opts
}

func newEditUserForm(users []apiclient.AdminUser) *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("user").Title("Utilisateur").Options(userOptions(users)...),
			huh.NewInput().Key("email").Title("Nouvel email (vide = inchangé)").Validate(optionalEmail),
			huh.NewInput().Key("password").Title("Nouveau mot de passe (vide = inchangé)").EchoMode(huh.EchoModePassword).Validate(optionalPassword),
			huh.NewConfirm().Key("ok").Title("Confirmer la modification ?"),
		),
	)
}

func newUserActionForm(users []apiclient.AdminUser) *huh.Form {
	return ui.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Key("user").Title("Utilisateur").Options(userOptions(users)...),
			huh.NewSelect[string]().Key("action").Title("Action").Options(
				huh.NewOption("Confirmer le compte", "confirm"),
				huh.NewOption("Réactiver", "enable"),
				huh.NewOption("Désactiver", "disable"),
				huh.NewOption("Passer admin", "role_admin"),
				huh.NewOption("Passer client", "role_user"),
				huh.NewOption("Supprimer", "delete"),
			),
			huh.NewConfirm().Key("ok").Title("Confirmer l'action ?"),
		),
	)
}
