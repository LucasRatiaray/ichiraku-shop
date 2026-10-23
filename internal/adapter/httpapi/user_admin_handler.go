package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/domain"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
)

type UserAdminHandler struct {
	users *usecase.UserAdminService
}

func NewUserAdminHandler(users *usecase.UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{users: users}
}

type adminUserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toAdminUserResponse(user domain.User) adminUserResponse {
	return adminUserResponse{ID: user.ID, Email: user.Email, Role: string(user.Role), Status: string(user.Status), CreatedAt: user.CreatedAt}
}

type adminCreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type adminUpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

func writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid email, password (8-72 chars), role or status")
	case errors.Is(err, domain.ErrSelfAction):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email already used")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "user has orders, disable it instead")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	default:
		log.Printf("admin users: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *UserAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		writeUserError(w, err)
		return
	}
	resp := make([]adminUserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, toAdminUserResponse(user))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *UserAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req adminCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	user, err := h.users.Create(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAdminUserResponse(user))
}

func (h *UserAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req adminUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := h.users.Update(r.Context(), actor.ID, id, req.Email, req.Password, req.Role, req.Status)
	if err != nil {
		writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAdminUserResponse(user))
}

func (h *UserAdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, _ := userFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.users.Delete(r.Context(), actor.ID, id); err != nil {
		writeUserError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
