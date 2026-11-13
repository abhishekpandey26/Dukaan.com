package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"gocommerce/middleware"
	"gocommerce/models"
	"gocommerce/utils"
)

type cartRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// GET /api/cart
func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())

	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT p.id, p.name, p.price_cents, p.image_url, p.stock, c.quantity
		FROM cart_items c
		JOIN products p ON p.id = c.product_id
		WHERE c.user_id = $1
		ORDER BY c.id`, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load cart")
		return
	}
	defer rows.Close()

	cart := models.Cart{Items: []models.CartItem{}}
	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&item.ProductID, &item.Name, &item.PriceCents, &item.ImageURL, &item.Stock, &item.Quantity); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "could not read cart")
			return
		}
		item.SubtotalCents = item.PriceCents * int64(item.Quantity)
		cart.TotalCents += item.SubtotalCents
		cart.Items = append(cart.Items, item)
	}
	utils.WriteJSON(w, http.StatusOK, cart)
}

// POST /api/cart   body: {"product_id": 1, "quantity": 2}
// Adds to the existing quantity if the product is already in the cart.
func (h *Handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())

	var req cartRequest
	if err := utils.ReadJSON(w, r, &req); err != nil || req.ProductID <= 0 || req.Quantity <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "product_id and a positive quantity are required")
		return
	}

	var stock, inCart int
	err := h.DB.QueryRowContext(r.Context(), `
		SELECT p.stock, COALESCE(c.quantity, 0)
		FROM products p
		LEFT JOIN cart_items c ON c.product_id = p.id AND c.user_id = $2
		WHERE p.id = $1`, req.ProductID, userID).Scan(&stock, &inCart)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not add to cart")
		return
	}
	// A friendly early check. The REAL guarantee happens at checkout inside a transaction.
	if inCart+req.Quantity > stock {
		utils.WriteError(w, http.StatusConflict, "not enough stock")
		return
	}

	// "Upsert": insert, or if (user_id, product_id) already exists, add to quantity.
	_, err = h.DB.ExecContext(r.Context(), `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`,
		userID, req.ProductID, req.Quantity)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not add to cart")
		return
	}
	h.GetCart(w, r) // reply with the updated cart
}

// PUT /api/cart/{productId}   body: {"quantity": 3}
func (h *Handler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())
	productID, ok := pathInt(r, "productId")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req cartRequest
	if err := utils.ReadJSON(w, r, &req); err != nil || req.Quantity <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "a positive quantity is required")
		return
	}

	var stock int
	err := h.DB.QueryRowContext(r.Context(), `SELECT stock FROM products WHERE id = $1`, productID).Scan(&stock)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if req.Quantity > stock {
		utils.WriteError(w, http.StatusConflict, "not enough stock")
		return
	}

	result, err := h.DB.ExecContext(r.Context(),
		`UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND product_id = $3`,
		req.Quantity, userID, productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not update cart")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		utils.WriteError(w, http.StatusNotFound, "item is not in your cart")
		return
	}
	h.GetCart(w, r)
}

// DELETE /api/cart/{productId}
func (h *Handler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())
	productID, ok := pathInt(r, "productId")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if _, err := h.DB.ExecContext(r.Context(),
		`DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`, userID, productID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not remove item")
		return
	}
	h.GetCart(w, r)
}

// DELETE /api/cart
func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())
	if _, err := h.DB.ExecContext(r.Context(), `DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not clear cart")
		return
	}
	h.GetCart(w, r)
}
