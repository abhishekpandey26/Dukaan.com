package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"gocommerce/models"
	"gocommerce/utils"
)

// All handlers in this file are mounted under /api/admin and wrapped in
// middleware.Admin, so only users with role "admin" can reach them.

// ---------- Products CRUD ----------

// POST /api/admin/products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var in models.ProductInput
	if err := utils.ReadJSON(w, r, &in); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := in.Validate(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	row := h.DB.QueryRowContext(r.Context(), `
		INSERT INTO products (name, description, price_cents, stock, category, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+productColumns,
		in.Name, in.Description, in.PriceCents, in.Stock, in.Category, in.ImageURL)
	product, err := scanProduct(row)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not create product")
		return
	}

	h.invalidateProductCache(r.Context())
	utils.WriteJSON(w, http.StatusCreated, product)
}

// PUT /api/admin/products/{id}
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var in models.ProductInput
	if err := utils.ReadJSON(w, r, &in); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := in.Validate(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	row := h.DB.QueryRowContext(r.Context(), `
		UPDATE products
		SET name = $1, description = $2, price_cents = $3, stock = $4,
		    category = $5, image_url = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING `+productColumns,
		in.Name, in.Description, in.PriceCents, in.Stock, in.Category, in.ImageURL, id)
	product, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not update product")
		return
	}

	h.invalidateProductCache(r.Context())
	utils.WriteJSON(w, http.StatusOK, product)
}

// DELETE /api/admin/products/{id}
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	result, err := h.DB.ExecContext(r.Context(), `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not delete product")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	h.invalidateProductCache(r.Context())
	w.WriteHeader(http.StatusNoContent) // 204 = success, no body
}

// ---------- Orders ----------

// GET /api/admin/orders?status=pending
func (h *Handler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var orders []models.Order
	var err error
	if status != "" {
		orders, err = h.fetchOrders(r.Context(), "WHERE o.status = $1", status)
	} else {
		orders, err = h.fetchOrders(r.Context(), "")
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load orders")
		return
	}
	utils.WriteJSON(w, http.StatusOK, orders)
}

// PUT /api/admin/orders/{id}/status   body: {"status": "shipped"}
func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	var body struct { // anonymous struct: handy for one-off request shapes
		Status string `json:"status"`
	}
	if err := utils.ReadJSON(w, r, &body); err != nil || !models.ValidOrderStatuses[body.Status] {
		utils.WriteError(w, http.StatusBadRequest, "status must be one of: pending, paid, shipped, delivered, cancelled")
		return
	}

	err := h.changeOrderStatus(r.Context(), id, body.Status, 0) // 0 = admin, any owner
	h.respondStatusChange(w, r, id, err)
}

// ---------- Users ----------

// GET /api/admin/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.QueryContext(r.Context(),
		`SELECT id, name, email, role, created_at FROM users ORDER BY id`)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load users")
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "could not read users")
			return
		}
		users = append(users, u)
	}
	utils.WriteJSON(w, http.StatusOK, users)
}

// ---------- Dashboard ----------

// GET /api/admin/stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	var s struct {
		Users        int   `json:"users"`
		Products     int   `json:"products"`
		Orders       int   `json:"orders"`
		PendingOrder int   `json:"pending_orders"`
		LowStock     int   `json:"low_stock"`
		RevenueCents int64 `json:"revenue_cents"`
	}
	err := h.DB.QueryRowContext(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM products),
			(SELECT COUNT(*) FROM orders),
			(SELECT COUNT(*) FROM orders WHERE status = 'pending'),
			(SELECT COUNT(*) FROM products WHERE stock < 5),
			(SELECT COALESCE(SUM(total_cents), 0) FROM orders WHERE status <> 'cancelled')
	`).Scan(&s.Users, &s.Products, &s.Orders, &s.PendingOrder, &s.LowStock, &s.RevenueCents)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	utils.WriteJSON(w, http.StatusOK, s)
}
