package admin

import (
	"context"
	"errors"
	"strconv"

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

type usersResultMsg struct {
	users []apiclient.AdminUser
	err   error
}

func listUsersCmd(api *apiclient.Client, token string) tea.Cmd {
	return func() tea.Msg {
		users, err := api.ListUsers(context.Background(), token)
		return usersResultMsg{users: users, err: err}
	}
}

type userActionMsg struct {
	info string
	err  error
}

func createUserCmd(api *apiclient.Client, token, email, password, role string) tea.Cmd {
	return func() tea.Msg {
		user, err := api.CreateUser(context.Background(), token, email, password, role)
		return userActionMsg{info: "Compte " + user.Email + " créé.", err: err}
	}
}

func editUserCmd(api *apiclient.Client, token, idStr, email, password string) tea.Cmd {
	return func() tea.Msg {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return userActionMsg{err: err}
		}
		_, err = api.UpdateUser(context.Background(), token, id, email, password, "", "")
		return userActionMsg{info: "Compte modifié.", err: err}
	}
}

func userActionCmd(api *apiclient.Client, token, idStr, action string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return userActionMsg{err: err}
		}
		switch action {
		case "delete":
			return userActionMsg{info: "Compte supprimé.", err: api.DeleteUser(ctx, token, id)}
		case "confirm", "enable":
			_, err = api.UpdateUser(ctx, token, id, "", "", "", "active")
			return userActionMsg{info: "Compte activé.", err: err}
		case "disable":
			_, err = api.UpdateUser(ctx, token, id, "", "", "", "disabled")
			return userActionMsg{info: "Compte désactivé.", err: err}
		case "role_admin":
			_, err = api.UpdateUser(ctx, token, id, "", "", "admin", "")
			return userActionMsg{info: "Rôle admin attribué.", err: err}
		case "role_user":
			_, err = api.UpdateUser(ctx, token, id, "", "", "user", "")
			return userActionMsg{info: "Rôle client attribué.", err: err}
		}
		return userActionMsg{info: "Aucune action."}
	}
}
