package admin

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/tui/ui"
)

type screen int

const (
	screenLogin screen = iota
	screenHome
	screenCreateProduct
	screenOrders
	screenChangeStatus
	screenCreateOrder
	screenUsers
	screenCreateUser
	screenUserAction
)

type Model struct {
	api      *apiclient.Client
	form     *huh.Form
	screen   screen
	loading  bool
	token    string
	err      string
	info     string
	quitting bool
	spinner  spinner.Model
	orders   []apiclient.Order
	products []apiclient.Product
	users    []apiclient.AdminUser
}

func New(api *apiclient.Client) Model {
	return Model{
		api:     api,
		screen:  screenLogin,
		form:    newLoginForm(),
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) startLoading(cmd tea.Cmd) (Model, tea.Cmd) {
	m.loading = true
	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m Model) goHome() (Model, tea.Cmd) {
	m.screen = screenHome
	m.form = newHomeForm()
	return m, m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		if (m.screen == screenCreateProduct || m.screen == screenCreateOrder) && msg.String() == "esc" {
			m.err = ""
			return m.goHome()
		}
		if m.screen == screenChangeStatus && msg.String() == "esc" {
			m.err = ""
			m.screen = screenOrders
			return m, nil
		}
		if (m.screen == screenCreateUser || m.screen == screenUserAction) && msg.String() == "esc" {
			m.err = ""
			m.screen = screenUsers
			return m, nil
		}
		if m.screen == screenOrders {
			switch msg.String() {
			case "s":
				if len(m.orders) == 0 {
					return m, nil
				}
				m.err = ""
				m.info = ""
				m.screen = screenChangeStatus
				m.form = newChangeStatusForm(m.orders)
				return m, m.form.Init()
			case "r":
				m.info = ""
				return m.startLoading(listAllOrdersCmd(m.api, m.token))
			case "esc", "q":
				m.err = ""
				m.info = ""
				return m.goHome()
			}
			return m, nil
		}
		if m.screen == screenUsers {
			switch msg.String() {
			case "c":
				m.err = ""
				m.info = ""
				m.screen = screenCreateUser
				m.form = newCreateUserForm()
				return m, m.form.Init()
			case "a":
				if len(m.users) == 0 {
					return m, nil
				}
				m.err = ""
				m.info = ""
				m.screen = screenUserAction
				m.form = newUserActionForm(m.users)
				return m, m.form.Init()
			case "r":
				m.info = ""
				return m.startLoading(listUsersCmd(m.api, m.token))
			case "esc", "q":
				m.err = ""
				m.info = ""
				return m.goHome()
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
		return m.goHome()

	case logoutResultMsg:
		if m.quitting {
			return m, tea.Quit
		}
		m.loading = false
		m.token = ""
		m.orders = nil
		m.products = nil
		m.users = nil
		m.info = "Tu es déconnecté."
		m.screen = screenLogin
		m.form = newLoginForm()
		return m, m.form.Init()

	case createProductResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newCreateProductForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.info = fmt.Sprintf("Produit %s créé (%s TTC)", msg.product.ID, ui.FormatEuros(msg.product.PriceTTC))
		return m.goHome()

	case ordersResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m.goHome()
		}
		m.err = ""
		m.orders = msg.orders
		m.screen = screenOrders
		return m, nil

	case statusChangedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newChangeStatusForm(m.orders)
			return m, m.form.Init()
		}
		m.err = ""
		m.info = fmt.Sprintf("Commande %s : %s", msg.order.ID, ui.OrderStatusLabel(msg.order.Status))
		return m.startLoading(listAllOrdersCmd(m.api, m.token))

	case productsResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m.goHome()
		}
		if len(msg.products) == 0 {
			m.err = "Aucun produit : crée d'abord un produit."
			return m.goHome()
		}
		m.err = ""
		m.products = msg.products
		m.screen = screenCreateOrder
		m.form = newCreateOrderForm(m.products)
		return m, m.form.Init()

	case orderCreatedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newCreateOrderForm(m.products)
			return m, m.form.Init()
		}
		m.err = ""
		m.info = fmt.Sprintf("Commande %s créée pour %s (en attente)", msg.order.ID, msg.order.UserEmail)
		return m.goHome()

	case usersResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m.goHome()
		}
		m.err = ""
		m.users = msg.users
		m.screen = screenUsers
		return m, nil

	case userActionMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.info = ""
			m.screen = screenUsers
			return m.startLoading(listUsersCmd(m.api, m.token))
		}
		m.err = ""
		m.info = msg.info
		return m.startLoading(listUsersCmd(m.api, m.token))
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
	case screenLogin:
		m.info = ""
		return m.startLoading(loginCmd(m.api, m.form.GetString("email"), m.form.GetString("password")))

	case screenHome:
		m.err = ""
		m.info = ""
		switch m.form.GetString("choice") {
		case "create_product":
			m.screen = screenCreateProduct
			m.form = newCreateProductForm()
			return m, m.form.Init()
		case "orders":
			return m.startLoading(listAllOrdersCmd(m.api, m.token))
		case "create_order":
			return m.startLoading(loadProductsCmd(m.api))
		case "users":
			return m.startLoading(listUsersCmd(m.api, m.token))
		case "logout":
			return m.startLoading(logoutCmd(m.api, m.token))
		case "quit":
			m.quitting = true
			return m.startLoading(logoutCmd(m.api, m.token))
		}

	case screenCreateProduct:
		price, _ := ui.ParseEuros(m.form.GetString("price"))
		return m.startLoading(createProductCmd(m.api, m.token,
			m.form.GetString("name"), m.form.GetString("description"), m.form.GetString("category"), price))

	case screenChangeStatus:
		return m.startLoading(changeStatusCmd(m.api, m.token,
			m.form.GetString("order"), m.form.GetString("status"), m.form.GetString("reason")))

	case screenCreateOrder:
		ids, _ := m.form.Get("products").([]string)
		qty, _ := strconv.Atoi(m.form.GetString("quantity"))
		return m.startLoading(createOrderCmd(m.api, m.token, m.form.GetString("email"), ids, qty))

	case screenCreateUser:
		return m.startLoading(createUserCmd(m.api, m.token,
			m.form.GetString("email"), m.form.GetString("password"), m.form.GetString("role")))

	case screenUserAction:
		if !m.form.GetBool("ok") {
			m.screen = screenUsers
			return m, nil
		}
		return m.startLoading(userActionCmd(m.api, m.token, m.form.GetString("user"), m.form.GetString("action")))
	}

	return m, cmd
}

func (m Model) View() tea.View {
	s := ui.TitleStyle.Render("🍜 Ichiraku Shop — Admin") + "\n\n"

	if m.loading {
		s += m.spinner.View() + " Chargement…\n"
		return tea.NewView(s)
	}

	help := "↑/↓ naviguer • entrée valider • ctrl+c quitter"
	switch m.screen {
	case screenOrders:
		s += renderOrders(m.orders)
		help = "s changer un statut • r rafraîchir • esc retour • ctrl+c quitter"
	case screenUsers:
		s += renderUsers(m.users)
		help = "c créer • a action (confirmer, rôle, désactiver, supprimer) • r rafraîchir • esc retour"
	default:
		s += m.form.View()
		switch m.screen {
		case screenCreateProduct, screenChangeStatus, screenCreateUser, screenUserAction:
			help = "↑/↓ naviguer • entrée valider • esc retour • ctrl+c quitter"
		case screenCreateOrder:
			help = "↑/↓ naviguer • espace cocher • entrée valider • esc retour • ctrl+c quitter"
		}
	}

	if m.err != "" {
		s += "\n" + ui.ErrStyle.Render("✗ "+m.err) + "\n"
	}
	if m.info != "" {
		s += "\n" + ui.InfoStyle.Render("✓ "+m.info) + "\n"
	}

	s += "\n" + ui.HelpStyle.Render(help)
	return tea.NewView(s)
}
