package client

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
)

type screen int

const (
	screenMenu screen = iota
	screenLogin
	screenRegister
	screenConfirm
	screenForgot
	screenReset
	screenHome
	screenSearch
	screenResults
)

type Model struct {
	api      *apiclient.Client
	form     *huh.Form
	screen   screen
	loading  bool
	token    string
	email    string
	err      string
	info     string
	quitting bool
	spinner  spinner.Model
	products []apiclient.Product
}

func New(api *apiclient.Client) Model {
	return Model{
		api:     api,
		screen:  screenMenu,
		form:    newMenuForm(),
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
		if m.screen == screenSearch && msg.String() == "esc" {
			m.err = ""
			m.screen = screenHome
			m.form = newHomeForm()
			return m, m.form.Init()
		}
		if m.screen == screenResults {
			switch msg.String() {
			case "n":
				m.screen = screenSearch
				m.form = newSearchForm()
				return m, m.form.Init()
			case "esc", "q":
				m.err = ""
				m.screen = screenHome
				m.form = newHomeForm()
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
		m.screen = screenHome
		m.form = newHomeForm()
		return m, m.form.Init()

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
		m.err = ""
		m.info = "Tu es déconnecté."
		m.screen = screenMenu
		m.form = newMenuForm()
		return m, m.form.Init()

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newSearchForm()
			return m, m.form.Init()
		}
		m.err = ""
		m.products = msg.products
		m.screen = screenResults
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
		case "forgot":
			m.screen = screenForgot
			m.form = newForgotForm()
			return m, m.form.Init()
		case "quit":
			return m, tea.Quit
		}

	case screenLogin:
		return m.startLoading(loginCmd(m.api, m.form.GetString("email"), m.form.GetString("password")))

	case screenRegister:
		m.email = m.form.GetString("email")
		return m.startLoading(registerCmd(m.api, m.email, m.form.GetString("password")))

	case screenConfirm:
		return m.startLoading(confirmCmd(m.api, m.email, m.form.GetString("code")))

	case screenForgot:
		m.email = m.form.GetString("email")
		return m.startLoading(forgotCmd(m.api, m.email))

	case screenReset:
		return m.startLoading(resetCmd(m.api, m.email, m.form.GetString("code"), m.form.GetString("password")))

	case screenHome:
		m.err = ""
		m.info = ""
		switch m.form.GetString("choice") {
		case "logout":
			return m.startLoading(logoutCmd(m.api, m.token))
		case "quit":
			m.quitting = true
			return m.startLoading(logoutCmd(m.api, m.token))
		case "catalog":
			m.screen = screenSearch
			m.form = newSearchForm()
			return m, m.form.Init()
		case "cart":
			m.info = "Bientôt disponible."
			m.form = newHomeForm()
			return m, m.form.Init()
		}
	case screenSearch:
		return m.startLoading(searchCmd(m.api, filterFromSearchForm(m.form)))
	}

	return m, cmd
}

func (m Model) View() tea.View {
	s := titleStyle.Render("🍜 Ichiraku Shop") + "\n\n"

	if m.loading {
		s += m.spinner.View() + " Chargement…\n"
		return tea.NewView(s)
	}

	help := "↑/↓ naviguer • entrée valider • ctrl+c quitter"
	if m.screen == screenResults {
		s += renderProducts(m.products)
		help = "n nouvelle recherche • esc retour • ctrl+c quitter"
	} else {
		s += m.form.View()
		if m.screen == screenSearch {
			help = "↑/↓ naviguer • entrée valider • esc retour • ctrl+c quitter"
		}
	}

	if m.err != "" {
		s += "\n" + errStyle.Render("✗ "+m.err) + "\n"
	}
	if m.info != "" {
		s += "\n" + infoStyle.Render("✓ "+m.info) + "\n"
	}

	s += "\n" + helpStyle.Render(help)
	return tea.NewView(s)
}
