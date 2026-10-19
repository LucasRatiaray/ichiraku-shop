package client

import (
	"context"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
)

type Model struct {
	api     *apiclient.Client
	form    *huh.Form
	loading bool
	token   string
	err     string
}

func New(api *apiclient.Client) Model {
	return Model{
		api:  api,
		form: newLoginForm(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
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
		m.err = ""
		return m, nil
	}

	if m.token != "" || m.loading {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		m.loading = true
		return m, loginCmd(m.api, m.form.GetString("email"), m.form.GetString("password"))
	}

	return m, cmd
}

func (m Model) View() tea.View {
	if m.token != "" {
		return tea.NewView("Connecté ✓\n\nctrl+c pour quitter.\n")
	}
	if m.loading {
		return tea.NewView("Connexion…\n")
	}

	s := m.form.View()
	if m.err != "" {
		s += "\n" + m.err + "\n"
	}
	return tea.NewView(s)
}

func newLoginForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("email").Title("Email"),
			huh.NewInput().Key("password").Title("Mot de passe").EchoMode(huh.EchoModePassword),
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