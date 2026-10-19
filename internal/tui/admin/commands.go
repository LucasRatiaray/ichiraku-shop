package admin

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/apiclient"
	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
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

type ordersResultMsg struct {
	orders []apiclient.Order
	err    error
}

func listAllOrdersCmd(api *apiclient.Client, token string) tea.Cmd {
	return func() tea.Msg {
		orders, err := api.ListAllOrders(context.Background(), token)
		return ordersResultMsg{orders: orders, err: err}
	}
}

type statusChangedMsg struct {
	order apiclient.Order
	err   error
}

func changeStatusCmd(api *apiclient.Client, token, orderID, status, reason string) tea.Cmd {
	return func() tea.Msg {
		o, err := api.ChangeOrderStatus(context.Background(), token, orderID, status, reason)
		return statusChangedMsg{order: o, err: err}
	}
}

type productsResultMsg struct {
	products []apiclient.Product
	err      error
}

func loadProductsCmd(api *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		products, err := api.SearchProducts(context.Background(), domain.ProductFilter{})
		return productsResultMsg{products: products, err: err}
	}
}

type orderCreatedMsg struct {
	order apiclient.Order
	err   error
}

func createOrderCmd(api *apiclient.Client, token, email string, productIDs []string, quantity int) tea.Cmd {
	return func() tea.Msg {
		lines := make([]apiclient.OrderLine, 0, len(productIDs))
		for _, id := range productIDs {
			lines = append(lines, apiclient.OrderLine{ProductID: id, Quantity: quantity})
		}
		o, err := api.CreateOrderForUser(context.Background(), token, email, lines)
		return orderCreatedMsg{order: o, err: err}
	}
}