package client

import (
	"context"

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
)

type Model struct {
	api     *apiclient.Client
	form    *huh.Form
	screen  screen
	loading bool
	token   string
	email   string
	err     string
	info    string
}

func New(api *apiclient.Client) Model {
	return Model{
		api:    api,
		screen: screenMenu,
		form:   newMenuForm(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.form = newLoginForm()
			return m, m.form.Init()
		}
		m.token = msg.token
		m.screen = screenHome
		m.err = ""
		m.info = ""
		return m, nil

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
	}

	if m.screen == screenHome || m.loading {
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
		m.loading = true
		return m, loginCmd(m.api, m.form.GetString("email"), m.form.GetString("password"))

	case screenRegister:
		m.email = m.form.GetString("email")
		m.loading = true
		return m, registerCmd(m.api, m.email, m.form.GetString("password"))

	case screenConfirm:
		m.loading = true
		return m, confirmCmd(m.api, m.email, m.form.GetString("code"))

	case screenForgot:
		m.email = m.form.GetString("email")
		m.loading = true
		return m, forgotCmd(m.api, m.email)

	case screenReset:
		m.loading = true
		return m, resetCmd(m.api, m.email, m.form.GetString("code"), m.form.GetString("password"))
	}

	return m, cmd
}

func (m Model) View() tea.View {
	if m.screen == screenHome {
		return tea.NewView("Connecté ✓\n\nctrl+c pour quitter.\n")
	}
	if m.loading {
		return tea.NewView("Chargement…\n")
	}

	s := m.form.View()
	if m.err != "" {
		s += "\n" + m.err + "\n"
	}
	if m.info != "" {
		s += "\n" + m.info + "\n"
	}
	return tea.NewView(s)
}

func newMenuForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("choice").
				Title("Ichiraku Shop").
				Options(
					huh.NewOption("Se connecter", "login"),
					huh.NewOption("Créer un compte", "register"),
					huh.NewOption("Mot de passe oublié", "forgot"),
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

type loginResultMsg struct {
	token string
	err   error
}

func loginCmd(api *apiclient.Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		token, err := api.Login(context.Background(), email, password)
		return loginResultMsg{token: token, err: err}
	}
}

type registerResultMsg struct {
	err error
}

func registerCmd(api *apiclient.Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		err := api.Register(context.Background(), email, password)
		return registerResultMsg{err: err}
	}
}

type confirmResultMsg struct {
	err error
}

func confirmCmd(api *apiclient.Client, email, code string) tea.Cmd {
	return func() tea.Msg {
		err := api.Confirm(context.Background(), email, code)
		return confirmResultMsg{err: err}
	}
}

type forgotResultMsg struct {
	err error
}

func forgotCmd(api *apiclient.Client, email string) tea.Cmd {
	return func() tea.Msg {
		err := api.ForgotPassword(context.Background(), email)
		return forgotResultMsg{err: err}
	}
}

type resetResultMsg struct {
	err error
}

func resetCmd(api *apiclient.Client, email, code, password string) tea.Cmd {
	return func() tea.Msg {
		err := api.ResetPassword(context.Background(), email, code, password)
		return resetResultMsg{err: err}
	}
}