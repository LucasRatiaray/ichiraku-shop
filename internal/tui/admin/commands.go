package admin

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
)

var errNotAdmin = errors.New("access restricted to admin users")

type loginResultMsg struct {
	token string
	err   error
}

func loginCmd(api *apiclient.Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		token, err := api.Login(ctx, email, password)
		if err != nil {
			return loginResultMsg{err: err}
		}
		u, err := api.Me(ctx, token)
		if err != nil {
			return loginResultMsg{err: err}
		}
		if u.Role != "admin" {
			_ = api.Logout(ctx, token)
			return loginResultMsg{err: errNotAdmin}
		}
		return loginResultMsg{token: token}
	}
}

type logoutResultMsg struct{ err error }

func logoutCmd(api *apiclient.Client, token string) tea.Cmd {
	return func() tea.Msg {
		return logoutResultMsg{err: api.Logout(context.Background(), token)}
	}
}

type createProductResultMsg struct {
	product apiclient.Product
	err     error
}

func createProductCmd(api *apiclient.Client, token, name, description, category string, priceHT int64) tea.Cmd {
	return func() tea.Msg {
		p, err := api.CreateProduct(context.Background(), token, name, description, category, priceHT)
		return createProductResultMsg{product: p, err: err}
	}
}
