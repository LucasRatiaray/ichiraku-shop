package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func (c *Client) do(ctx context.Context, method, path, token string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("api: %s %s: %w", method, path, err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("api: %s %s: %w", method, path, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("api: %s %s: %w", method, path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return &APIError{
			Status:  resp.StatusCode,
			Message: apiErr.Error,
		}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("api: %s %s: %w", method, path, err)
		}
	}

	return nil
}

func (c *Client) Register(ctx context.Context, email, password string) error {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	return c.do(ctx, http.MethodPost, "/auth/register", "", body, nil)
}

func (c *Client) Confirm(ctx context.Context, email, code string) error {
	body := map[string]string{
		"email": email,
		"code":  code,
	}
	return c.do(ctx, http.MethodPost, "/auth/confirm", "", body, nil)
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	var res struct {
		Token string `json:"token"`
	}

	if err := c.do(ctx, http.MethodPost, "/auth/login", "", body, &res); err != nil {
		return "", err
	}
	return res.Token, nil
}

func (c *Client) Logout(ctx context.Context, token string) error {
	return c.do(ctx, http.MethodPost, "/auth/logout", token, nil, nil)
}

func (c *Client) ForgotPassword(ctx context.Context, email string) error {
	body := map[string]string{"email": email}
	return c.do(ctx, http.MethodPost, "/auth/password/forgot", "", body, nil)
}

func (c *Client) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	body := map[string]string{
		"email":        email,
		"code":         code,
		"new_password": newPassword,
	}
	return c.do(ctx, http.MethodPost, "/auth/password/reset", "", body, nil)
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PriceHT     int64  `json:"price_ht"`
	PriceTTC    int64  `json:"price_ttc"`
}

func (c *Client) SearchProducts(ctx context.Context, f domain.ProductFilter) ([]Product, error) {
	q := url.Values{}
	if f.Query != "" {
		q.Set("q", f.Query)
	}
	if f.Category != "" {
		q.Set("category", f.Category)
	}
	setInt := func(key string, v int64) {
		if v > 0 {
			q.Set(key, strconv.FormatInt(v, 10))
		}
	}
	setInt("min_price", f.MinPriceHT)
	setInt("max_price", f.MaxPriceHT)
	setInt("min_ttc", f.MinPriceTTC)
	setInt("max_ttc", f.MaxPriceTTC)

	path := "/products"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var products []Product
	if err := c.do(ctx, http.MethodGet, path, "", nil, &products); err != nil {
		return nil, err
	}
	return products, nil
}

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (c *Client) Me(ctx context.Context, token string) (User, error) {
	var u User
	if err := c.do(ctx, http.MethodGet, "/auth/me", token, nil, &u); err != nil {
		return User{}, err
	}
	return u, nil
}

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PriceHT     int64  `json:"price_ht"`
}

func (c *Client) CreateProduct(ctx context.Context, token, name, description, category string, priceHT int64) (Product, error) {
	req := createProductRequest{
		Name:        name,
		Description: description,
		Category:    category,
		PriceHT:     priceHT,
	}
	var p Product
	if err := c.do(ctx, http.MethodPost, "/admin/products", token, req, &p); err != nil {
		return Product{}, err
	}
	return p, nil
}

type BasketItem struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	UnitPriceTTC int64  `json:"unit_price_ttc"`
	Quantity     int    `json:"quantity"`
	TotalTTC     int64  `json:"total_ttc"`
}

type Basket struct {
	ID           string       `json:"id"`
	Items        []BasketItem `json:"items"`
	ShippingCost int64        `json:"shipping_cost"`
	TotalTTC     int64        `json:"total_ttc"`
}

type addItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type setQuantityRequest struct {
	Quantity int `json:"quantity"`
}

func (c *Client) GetBasket(ctx context.Context, token string) (Basket, error) {
	var b Basket
	err := c.do(ctx, http.MethodGet, "/basket", token, nil, &b)
	return b, err
}

func (c *Client) AddToBasket(ctx context.Context, token, productID string, quantity int) (Basket, error) {
	var b Basket
	err := c.do(ctx, http.MethodPost, "/basket/items", token, addItemRequest{ProductID: productID, Quantity: quantity}, &b)
	return b, err
}

func (c *Client) SetBasketQuantity(ctx context.Context, token, productID string, quantity int) (Basket, error) {
	var b Basket
	err := c.do(ctx, http.MethodPut, "/basket/items/"+url.PathEscape(productID), token, setQuantityRequest{Quantity: quantity}, &b)
	return b, err
}

func (c *Client) RemoveFromBasket(ctx context.Context, token, productID string) (Basket, error) {
	var b Basket
	err := c.do(ctx, http.MethodDelete, "/basket/items/"+url.PathEscape(productID), token, nil, &b)
	return b, err
}

type OrderItem struct {
	ProductID    string `json:"product_id"`
	Name         string `json:"name"`
	UnitPriceTTC int64  `json:"unit_price_ttc"`
	Quantity     int    `json:"quantity"`
	TotalTTC     int64  `json:"total_ttc"`
}

type Order struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	CancelReason string      `json:"cancel_reason"`
	TotalTTC     int64       `json:"total_ttc"`
	Items        []OrderItem `json:"items"`
	CreatedAt    time.Time   `json:"created_at"`
	UserEmail    string      `json:"user_email"`
}

type checkoutRequest struct {
	CardNumber string `json:"card_number"`
	Expiry     string `json:"expiry"`
	CVC        string `json:"cvc"`
}

func (c *Client) Checkout(ctx context.Context, token, cardNumber, expiry, cvc string) (Order, error) {
	var o Order
	err := c.do(ctx, http.MethodPost, "/orders", token, checkoutRequest{CardNumber: cardNumber, Expiry: expiry, CVC: cvc}, &o)
	return o, err
}

func (c *Client) ListOrders(ctx context.Context, token string) ([]Order, error) {
	var orders []Order
	err := c.do(ctx, http.MethodGet, "/orders", token, nil, &orders)
	return orders, err
}

type changeStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type OrderLine struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type adminCreateOrderRequest struct {
	UserEmail string      `json:"user_email"`
	Items     []OrderLine `json:"items"`
}

func (c *Client) ListAllOrders(ctx context.Context, token string) ([]Order, error) {
	var orders []Order
	err := c.do(ctx, http.MethodGet, "/admin/orders", token, nil, &orders)
	return orders, err
}

func (c *Client) ChangeOrderStatus(ctx context.Context, token, orderID, status, reason string) (Order, error) {
	var o Order
	path := "/admin/orders/" + url.PathEscape(orderID) + "/status"
	err := c.do(ctx, http.MethodPatch, path, token, changeStatusRequest{Status: status, Reason: reason}, &o)
	return o, err
}

func (c *Client) CreateOrderForUser(ctx context.Context, token, email string, items []OrderLine) (Order, error) {
	var o Order
	err := c.do(ctx, http.MethodPost, "/admin/orders", token, adminCreateOrderRequest{UserEmail: email, Items: items}, &o)
	return o, err
}

type AdminUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type adminCreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type adminUpdateUserRequest struct {
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

func (c *Client) ListUsers(ctx context.Context, token string) ([]AdminUser, error) {
	var users []AdminUser
	err := c.do(ctx, http.MethodGet, "/admin/users", token, nil, &users)
	return users, err
}

func (c *Client) CreateUser(ctx context.Context, token, email, password, role string) (AdminUser, error) {
	var user AdminUser
	err := c.do(ctx, http.MethodPost, "/admin/users", token, adminCreateUserRequest{Email: email, Password: password, Role: role}, &user)
	return user, err
}

func (c *Client) UpdateUser(ctx context.Context, token string, id int64, role, status string) (AdminUser, error) {
	var user AdminUser
	path := "/admin/users/" + strconv.FormatInt(id, 10)
	err := c.do(ctx, http.MethodPatch, path, token, adminUpdateUserRequest{Role: role, Status: status}, &user)
	return user, err
}

func (c *Client) DeleteUser(ctx context.Context, token string, id int64) error {
	path := "/admin/users/" + strconv.FormatInt(id, 10)
	return c.do(ctx, http.MethodDelete, path, token, nil, nil)
}
