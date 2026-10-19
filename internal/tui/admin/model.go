package admin

import (
	"fmt"

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
		if m.screen == screenCreateProduct && msg.String() == "esc" {
			m.err = ""
			return m.goHome()
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
	}

	return m, cmd
}

func (m Model) View() tea.View {
	s := ui.TitleStyle.Render("🍜 Ichiraku Shop — Admin") + "\n\n"

	if m.loading {
		s += m.spinner.View() + " Chargement…\n"
		return tea.NewView(s)
	}

	s += m.form.View()
	if m.err != "" {
		s += "\n" + ui.ErrStyle.Render("✗ "+m.err) + "\n"
	}
	if m.info != "" {
		s += "\n" + ui.InfoStyle.Render("✓ "+m.info) + "\n"
	}

	help := "↑/↓ naviguer • entrée valider • ctrl+c quitter"
	if m.screen == screenCreateProduct {
		help = "↑/↓ naviguer • entrée valider • esc retour • ctrl+c quitter"
	}
	s += "\n" + ui.HelpStyle.Render(help)
	return tea.NewView(s)
}