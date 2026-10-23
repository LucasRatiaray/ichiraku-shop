package client

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

type screen int

const (
	screenMenu screen = iota
	screenLogin
	screenRegister
	screenConfirm
	screenResend
	screenForgot
	screenReset
	screenSearch
	screenResults
	screenAddToBasket
	screenBasket
	screenEditBasket
	screenPayment
	screenOrders
	screenAccount
	screenChangePassword
)

type Model struct {
	api      *apiclient.Client
	form     *huh.Form
	screen   screen
	width    int
	height   int
	loading  bool
	token    string
	email    string
	err      string
	info     string
	quitting bool
	spinner  spinner.Model
	products []apiclient.Product
	basket   apiclient.Basket
	orders   []apiclient.Order
}

func New(api *apiclient.Client) Model {
	return Model{
		api:     api,
		screen:  screenMenu,
		form:    newMenuForm(),
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(ui.TitleStyle)),
	}
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) startLoading(cmd tea.Cmd) (Model, tea.Cmd) {
	m.loading = true
	return m, tea.Batch(cmd, m.spinner.Tick) // lance le spinner et la commande en parallèle
}

func (m Model) loadCatalog() (Model, tea.Cmd) {
	return m.startLoading(searchCmd(m.api, domain.ProductFilter{}))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if !m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.screen == screenSearch && msg.String() == "esc" {
			m.err = ""
			m.screen = screenResults
			return m, nil
		}
		if m.screen == screenAddToBasket && msg.String() == "esc" {
			m.err = ""
			m.screen = screenResults
			return m, nil
		}
		if (m.screen == screenEditBasket || m.screen == screenPayment) && msg.String() == "esc" {
			m.err = ""
			m.screen = screenBasket
			return m, nil
		}
		if m.screen == screenChangePassword && msg.String() == "esc" {
			m.err = ""
			m.screen = screenAccount
			return m, nil
		}
		if !m.loading && msg.String() == "esc" {
			switch m.screen {
			case screenLogin, screenRegister, screenConfirm, screenResend, screenForgot, screenReset:
				m.err = ""
				m.info = ""
				m.screen = screenMenu
				m.form = newMenuForm()
				return m, m.form.Init()
			case screenBasket, screenOrders, screenAccount:
				m.err = ""
				m.info = ""
				m.screen = screenResults
				return m, nil
			}
		}
		if !m.loading && (m.screen == screenResults || m.screen == screenBasket || m.screen == screenOrders || m.screen == screenAccount) {
			switch msg.String() {
			case "u":
				m.err = ""
				m.info = ""
				m.screen = screenAccount
				return m, nil
			case "s":
				m.err = ""
				m.info = ""
				return m.loadCatalog()
			case "c":
				m.err = ""
				m.info = ""
				return m.startLoading(getBasketCmd(m.api, m.token))
			case "o":
				m.err = ""
				m.info = ""
				return m.startLoading(listOrdersCmd(m.api, m.token))
			case "q":
				m.quitting = true
				return m.startLoading(logoutCmd(m.api, m.token))
			}
		}
		if m.screen == screenResults {
			switch msg.String() {
			case "a":
				if len(m.products) == 0 {
					return m, nil
				}
				m.screen = screenAddToBasket
				m.form = newAddToBasketForm(m.products)
				return m, m.form.Init()
			case "/":
				m.screen = screenSearch
				m.form = newSearchForm()
				return m, m.form.Init()
			}
			return m, nil
		}
		if m.screen == screenBasket {
			switch msg.String() {
			case "p":
				if len(m.basket.Items) == 0 {
					m.err = "Ton panier est vide."
					return m, nil
				}
				m.err = ""
				m.info = ""
				m.screen = screenPayment
				m.form = newPaymentForm(m.basket.TotalTTC)
				return m, m.form.Init()
			case "m":
				if len(m.basket.Items) == 0 {
					return m, nil
				}
				m.info = ""
				m.screen = screenEditBasket
				m.form = newEditBasketForm(m.basket.Items)
				return m, m.form.Init()
			}
			return m, nil
		}
		if m.screen == screenOrders {
			return m, nil
		}
		if m.screen == screenAccount {
			if msg.String() == "p" {
				m.err = ""
				m.info = ""
				m.screen = screenChangePassword
				m.form = newChangePasswordForm()
				return m, m.form.Init()
			}
			return m, nil
		}

	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newLoginForm()
			return m, m.form.Init()
		}
		m.token = msg.token
		m.err = ""
		m.info = ""
		return m.startLoading(getBasketCmd(m.api, m.token))

	case registerResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newRegisterForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.screen = screenConfirm
		m.form = newConfirmForm()
		return m, m.form.Init()

	case confirmResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newConfirmForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.info = "Compte confirmé, tu peux te connecter."
		m.screen = screenLogin
		m.form = newLoginForm()
		return m, m.form.Init()

	case resendResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newForgotForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.info = "Si ce compte attend une confirmation, un nouveau code a été envoyé."
		m.screen = screenConfirm
		m.form = newConfirmForm()
		return m, m.form.Init()

	case forgotResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newForgotForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.info = "Si ce compte existe, un code a été envoyé."
		m.screen = screenReset
		m.form = newResetForm()
		return m, m.form.Init()

	case resetResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newResetForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.info = "Mot de passe modifié, tu peux te connecter."
		m.screen = screenLogin
		m.form = newLoginForm()
		return m, m.form.Init()

	case logoutResultMsg:
		if m.quitting {
			return m, tea.Quit
		}
		m.loading = false
		m.token = ""
		m.basket = apiclient.Basket{}
		m.products = nil
		m.orders = nil
		m.err = ""
		m.info = "Tu es déconnecté."
		m.screen = screenMenu
		m.form = newMenuForm()
		return m, m.form.Init()

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.screen = screenSearch
			m.form = newSearchForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.products = msg.products
		m.screen = screenResults
		return m, nil

	case basketResultMsg:
		m.loading = false
		if m.screen == screenLogin {
			m.basket = msg.basket
			return m.loadCatalog()
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.info = ""
			switch m.screen {
			case screenAddToBasket:
				m.form = newAddToBasketForm(m.products)
				return m, m.form.Init()
			case screenEditBasket:
				m.form = newEditBasketForm(m.basket.Items)
				return m, m.form.Init()
			}
			m.screen = screenResults
			return m, nil
		}
		m.err = ""
		m.basket = msg.basket
		m.screen = screenBasket
		return m, nil

	case orderResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newPaymentForm(m.basket.TotalTTC)
			return m, m.form.Init()
		}
		m.err = ""
		m.basket = apiclient.Basket{}
		m.info = fmt.Sprintf("Commande %s payée — %s TTC", msg.order.ID, ui.FormatEuros(msg.order.TotalTTC))
		return m.loadCatalog()

	case changePasswordResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newChangePasswordForm()
			return m, m.form.Init()
		}
		m.token = msg.token
		m.err = ""
		m.info = "Mot de passe modifié."
		m.screen = screenAccount
		return m, nil

	case ordersResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.screen = screenResults
			return m, nil
		}
		m.err = ""
		m.orders = msg.orders
		m.screen = screenOrders
		return m, nil
	}

	if m.loading {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State != huh.StateCompleted {
		return m, cmd
	}

	switch m.screen {
	case screenMenu:
		m.err = ""
		m.info = ""
		switch m.form.GetString("choice") {
		case "login":
			m.screen = screenLogin
			m.form = newLoginForm()
			return m, m.form.Init()
		case "register":
			m.screen = screenRegister
			m.form = newRegisterForm()
			return m, m.form.Init()
		case "resend":
			m.screen = screenResend
			m.form = newForgotForm()
			return m, m.form.Init()
		case "forgot":
			m.screen = screenForgot
			m.form = newForgotForm()
			return m, m.form.Init()
		case "quit":
			return m, tea.Quit
		}

	case screenLogin:
		m.email = strings.ToLower(strings.TrimSpace(m.form.GetString("email")))
		return m.startLoading(loginCmd(m.api, m.form.GetString("email"), m.form.GetString("password")))

	case screenRegister:
		m.email = m.form.GetString("email")
		return m.startLoading(registerCmd(m.api, m.email, m.form.GetString("password")))

	case screenConfirm:
		return m.startLoading(confirmCmd(m.api, m.email, m.form.GetString("code")))

	case screenResend:
		m.email = m.form.GetString("email")
		return m.startLoading(resendCmd(m.api, m.email))

	case screenForgot:
		m.email = m.form.GetString("email")
		return m.startLoading(forgotCmd(m.api, m.email))

	case screenReset:
		return m.startLoading(resetCmd(m.api, m.email, m.form.GetString("code"), m.form.GetString("password")))

	case screenSearch:
		return m.startLoading(searchCmd(m.api, filterFromSearchForm(m.form)))

	case screenAddToBasket:
		qty, _ := strconv.Atoi(m.form.GetString("quantity"))
		m.info = "Ajouté au panier."
		return m.startLoading(addToBasketCmd(m.api, m.token, m.form.GetString("product"), qty))

	case screenEditBasket:
		qty, _ := strconv.Atoi(m.form.GetString("quantity"))
		m.info = "Panier mis à jour."
		return m.startLoading(setBasketQuantityCmd(m.api, m.token, m.form.GetString("product"), qty))

	case screenPayment:
		return m.startLoading(checkoutCmd(m.api, m.token,
			m.form.GetString("card_number"), m.form.GetString("expiry"), m.form.GetString("cvc")))

	case screenChangePassword:
		return m.startLoading(changePasswordCmd(m.api, m.token,
			m.form.GetString("current"), m.form.GetString("new")))
	}

	return m, cmd
}

func (s screen) title() string {
	switch s {
	case screenMenu:
		return "Bienvenue"
	case screenLogin:
		return "Connexion"
	case screenRegister:
		return "Créer un compte"
	case screenConfirm:
		return "Confirmation du compte"
	case screenResend:
		return "Renvoyer le code"
	case screenForgot:
		return "Mot de passe oublié"
	case screenReset:
		return "Nouveau mot de passe"
	case screenSearch:
		return "Recherche"
	case screenResults:
		return "Catalogue"
	case screenAddToBasket:
		return "Ajouter au panier"
	case screenBasket:
		return "Panier"
	case screenEditBasket:
		return "Modifier le panier"
	case screenPayment:
		return "Paiement"
	case screenOrders:
		return "Mes commandes"
	case screenAccount:
		return "Compte"
	case screenChangePassword:
		return "Changer le mot de passe"
	}
	return ""
}

func (m Model) tabs() []ui.Tab {
	count := 0
	for _, it := range m.basket.Items {
		count += it.Quantity
	}
	return []ui.Tab{
		{Key: "s", Label: "boutique", Active: m.screen == screenResults || m.screen == screenSearch || m.screen == screenAddToBasket},
		{Key: "c", Label: "panier", Detail: fmt.Sprintf("%s [%d]", ui.FormatEuros(m.basket.TotalTTC), count),
			Active: m.screen == screenBasket || m.screen == screenEditBasket || m.screen == screenPayment},
		{Key: "o", Label: "commandes", Active: m.screen == screenOrders},
		{Key: "u", Label: "compte", Active: m.screen == screenAccount || m.screen == screenChangePassword},
		{Key: "q", Label: "quitter"},
	}
}

func (m Model) View() tea.View {
	contentWidth := ui.ContentWidth(m.width)
	quit := ui.Hint{Key: "ctrl+c", Label: "quitter"}
	hints := []ui.Hint{{Key: "↑/↓", Label: "naviguer"}, {Key: "entrée", Label: "valider"}, {Key: "esc", Label: "retour"}, quit}
	var s string

	switch m.screen {
	case screenResults:
		s = renderProducts(m.products, contentWidth)
		hints = []ui.Hint{{Key: "/", Label: "rechercher"}, {Key: "a", Label: "ajouter au panier"}, quit}
	case screenBasket:
		s = renderBasket(m.basket, contentWidth)
		hints = []ui.Hint{{Key: "p", Label: "payer"}, {Key: "m", Label: "modifier une quantité"}, {Key: "esc", Label: "retour"}, quit}
	case screenOrders:
		s = renderOrders(m.orders, contentWidth)
		hints = []ui.Hint{{Key: "esc", Label: "retour"}, quit}
	case screenAccount:
		s = renderAccount(m.email, contentWidth)
		hints = []ui.Hint{{Key: "p", Label: "changer le mot de passe"}, {Key: "esc", Label: "retour"}, quit}
	default:
		s = m.form.WithWidth(contentWidth).View()
	}
	if m.loading {
		s = m.spinner.View() + " Chargement…"
	}

	if !m.loading && m.err != "" {
		s += "\n" + ui.ErrStyle.Render("✗ "+m.err) + "\n"
	}
	if !m.loading && m.info != "" {
		s += "\n" + ui.InfoStyle.Render("✓ "+m.info) + "\n"
	}

	if m.token != "" {
		return ui.Shop(m.width, m.height, "Ichiraku Shop", m.tabs(), s, hints...)
	}
	help := "↑/↓ naviguer • entrée valider • ctrl+c quitter"
	if m.screen != screenMenu {
		help = "↑/↓ naviguer • entrée valider • esc retour • ctrl+c quitter"
	}
	if m.loading {
		help = "ctrl+c quitter"
	}
	return ui.Page(m.width, m.height, "Ichiraku Shop", m.screen.title(), s, help)
}
