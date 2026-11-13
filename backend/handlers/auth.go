package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/lib/pq"

	"gocommerce/auth"
	"gocommerce/middleware"
	"gocommerce/models"
	"gocommerce/utils"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := utils.ReadJSON(w, r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || !strings.Contains(req.Email, "@") || len(req.Password) < 6 {
		utils.WriteError(w, http.StatusBadRequest, "name, valid email and password (min 6 chars) are required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	var user models.User
	err = h.DB.QueryRowContext(r.Context(), `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, role, created_at`,
		req.Name, req.Email, hash,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		// errors.As checks if err is (or wraps) a *pq.Error. 23505 = unique_violation.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			utils.WriteError(w, http.StatusConflict, "email is already registered")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	h.respondWithToken(w, http.StatusCreated, user)
}

// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := utils.ReadJSON(w, r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var user models.User
	err := h.DB.QueryRowContext(r.Context(), `
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(req.Email)),
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

	// Same message for "no such user" and "wrong password" so attackers
	// can't find out which emails are registered.
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !auth.CheckPassword(user.PasswordHash, req.Password)) {
		utils.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	h.respondWithToken(w, http.StatusOK, user)
}

// GET /api/auth/me  (requires login)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())

	var user models.User
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT id, name, email, role, created_at FROM users WHERE id = $1`, userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) respondWithToken(w http.ResponseWriter, status int, user models.User) {
	token, err := auth.GenerateToken(h.JWTSecret, user.ID, user.Role)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	utils.WriteJSON(w, status, authResponse{Token: token, User: user})
}
