package client

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

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

type logoutResultMsg struct {
	err error
}

func logoutCmd(api *apiclient.Client, token string) tea.Cmd {
	return func() tea.Msg {
		err := api.Logout(context.Background(), token)
		return logoutResultMsg{err: err}
	}
}

type searchResultMsg struct {
	products []apiclient.Product
	err      error
}

func searchCmd(api *apiclient.Client, f domain.ProductFilter) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts(context.Background(), f)
		return searchResultMsg{products: products, err: err}
	}
}

type basketResultMsg struct {
	basket apiclient.Basket
	err    error
}

func getBasketCmd(api *apiclient.Client, token string) tea.Cmd {
	return func() tea.Msg {
		b, err := api.GetBasket(context.Background(), token)
		return basketResultMsg{basket: b, err: err}
	}
}

func addToBasketCmd(api *apiclient.Client, token, productID string, quantity int) tea.Cmd {
	return func() tea.Msg {
		b, err := api.AddToBasket(context.Background(), token, productID, quantity)
		return basketResultMsg{basket: b, err: err}
	}
}

func setBasketQuantityCmd(api *apiclient.Client, token, productID string, quantity int) tea.Cmd {
	return func() tea.Msg {
		b, err := api.SetBasketQuantity(context.Background(), token, productID, quantity)
		return basketResultMsg{basket: b, err: err}
	}
}
