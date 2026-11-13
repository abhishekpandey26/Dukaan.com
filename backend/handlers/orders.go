package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"

	"gocommerce/middleware"
	"gocommerce/models"
	"gocommerce/utils"
)

// ---------- Errors ----------
// "Sentinel errors" are package-level error values you compare with errors.Is.
var (
	ErrEmptyCart        = errors.New("your cart is empty")
	ErrOrderNotFound    = errors.New("order not found")
	ErrStatusNotAllowed = errors.New("status change not allowed")
)

// InsufficientStockError is a custom error TYPE carrying extra data.
// Any type with an Error() string method satisfies the built-in `error` interface.
type InsufficientStockError struct {
	ProductName string
	Available   int
	Requested   int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("not enough stock for %q: requested %d, only %d left", e.ProductName, e.Requested, e.Available)
}

// ---------- Place order (checkout) ----------

// POST /api/orders  -> turns the user's cart into an order.
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())

	orderID, err := h.placeOrder(r.Context(), userID)

	var stockErr *InsufficientStockError
	switch {
	case errors.Is(err, ErrEmptyCart):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	case errors.As(err, &stockErr):
		utils.WriteError(w, http.StatusConflict, stockErr.Error())
		return
	case err != nil:
		utils.WriteError(w, http.StatusInternalServerError, "could not place order")
		return
	}

	orders, err := h.fetchOrders(r.Context(), "WHERE o.id = $1", orderID)
	if err != nil || len(orders) == 0 {
		utils.WriteError(w, http.StatusInternalServerError, "order placed but could not be loaded")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, orders[0])
}

// placeOrder is the heart of the project. It runs inside a DATABASE TRANSACTION:
// either ALL of these steps happen, or NONE of them do.
//
//  1. lock the products in the cart (SELECT ... FOR UPDATE)
//  2. check there is enough stock
//  3. create the order + order items
//  4. reduce the stock
//  5. empty the cart
//
// Why lock? Imagine 2 users buying the last item at the same moment. Without the
// lock both could read "stock = 1" and both orders succeed -> stock becomes -1.
// FOR UPDATE makes the second transaction WAIT until the first one finishes,
// then it sees "stock = 0" and fails correctly.
func (h *Handler) placeOrder(ctx context.Context, userID int) (int, error) {
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	// If we return early with an error, Rollback undoes everything.
	// After a successful Commit, Rollback does nothing. So this is always safe.
	defer tx.Rollback()

	// Step 1: read cart + lock product rows. ORDER BY p.id makes every transaction
	// lock rows in the same order, which prevents deadlocks.
	rows, err := tx.QueryContext(ctx, `
		SELECT p.id, p.name, p.price_cents, p.stock, c.quantity
		FROM cart_items c
		JOIN products p ON p.id = c.product_id
		WHERE c.user_id = $1
		ORDER BY p.id
		FOR UPDATE OF p`, userID)
	if err != nil {
		return 0, err
	}

	type line struct {
		productID, stock, quantity int
		name                       string
		price                      int64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.productID, &l.name, &l.price, &l.stock, &l.quantity); err != nil {
			rows.Close()
			return 0, err
		}
		lines = append(lines, l)
	}
	rows.Close() // must close before running the next query in the same transaction
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, ErrEmptyCart
	}

	// Step 2: check stock and compute total.
	var total int64
	for _, l := range lines {
		if l.quantity > l.stock {
			return 0, &InsufficientStockError{ProductName: l.name, Available: l.stock, Requested: l.quantity}
		}
		total += l.price * int64(l.quantity)
	}

	// Step 3: create the order.
	var orderID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, total_cents) VALUES ($1, $2) RETURNING id`,
		userID, total).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	for _, l := range lines {
		// Step 3b: order item (copy of name & price at time of purchase).
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, price_cents, quantity)
			VALUES ($1, $2, $3, $4, $5)`,
			orderID, l.productID, l.name, l.price, l.quantity); err != nil {
			return 0, err
		}
		// Step 4: reduce stock.
		if _, err := tx.ExecContext(ctx,
			`UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2`,
			l.quantity, l.productID); err != nil {
			return 0, err
		}
	}

	// Step 5: empty the cart.
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	h.invalidateProductCache(ctx) // stock changed, so cached products are stale
	return orderID, nil
}

// ---------- Read orders ----------

// GET /api/orders  -> the logged-in user's orders
func (h *Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.fetchOrders(r.Context(), "WHERE o.user_id = $1", middleware.UserID(r.Context()))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load orders")
		return
	}
	utils.WriteJSON(w, http.StatusOK, orders)
}

// GET /api/orders/{id}
func (h *Handler) GetMyOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	orders, err := h.fetchOrders(r.Context(), "WHERE o.id = $1 AND o.user_id = $2", id, middleware.UserID(r.Context()))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load order")
		return
	}
	if len(orders) == 0 {
		utils.WriteError(w, http.StatusNotFound, "order not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, orders[0])
}

// POST /api/orders/{id}/cancel  -> customers can cancel their own PENDING orders.
func (h *Handler) CancelMyOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	err := h.changeOrderStatus(r.Context(), id, "cancelled", middleware.UserID(r.Context()))
	h.respondStatusChange(w, r, id, err)
}

// fetchOrders loads orders (filtered by `where`) together with their items.
// It uses 2 queries in total instead of 1 query per order (avoids the "N+1 problem").
func (h *Handler) fetchOrders(ctx context.Context, where string, args ...any) ([]models.Order, error) {
	rows, err := h.DB.QueryContext(ctx, `
		SELECT o.id, o.user_id, u.email, o.total_cents, o.status, o.created_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		`+where+`
		ORDER BY o.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.Order{}
	positionByID := map[int]int{} // order id -> index in the orders slice
	var ids []int64
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.UserEmail, &o.TotalCents, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{}
		positionByID[o.ID] = len(orders)
		orders = append(orders, o)
		ids = append(ids, int64(o.ID))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return orders, nil
	}

	// "= ANY($1)" with pq.Array lets us pass a whole Go slice as one parameter.
	itemRows, err := h.DB.QueryContext(ctx, `
		SELECT order_id, product_id, product_name, price_cents, quantity
		FROM order_items WHERE order_id = ANY($1) ORDER BY id`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var orderID int
		var productID sql.NullInt64 // NullInt64 can hold a NULL from the database
		var item models.OrderItem
		if err := itemRows.Scan(&orderID, &productID, &item.ProductName, &item.PriceCents, &item.Quantity); err != nil {
			return nil, err
		}
		if productID.Valid {
			pid := int(productID.Int64)
			item.ProductID = &pid
		}
		pos := positionByID[orderID]
		orders[pos].Items = append(orders[pos].Items, item)
	}
	return orders, itemRows.Err()
}

// ---------- Change status (shared by customer cancel + admin) ----------

// changeOrderStatus updates an order's status in a transaction.
// When an order is cancelled, its items go back into stock.
// ownerID = 0 means "admin": can change any order to any status.
func (h *Handler) changeOrderStatus(ctx context.Context, orderID int, newStatus string, ownerID int) error {
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var current string
	var userID int
	err = tx.QueryRowContext(ctx,
		`SELECT status, user_id FROM orders WHERE id = $1 FOR UPDATE`, orderID).Scan(&current, &userID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && ownerID != 0 && userID != ownerID) {
		return ErrOrderNotFound // don't reveal that other users' orders exist
	}
	if err != nil {
		return err
	}

	switch {
	case current == newStatus:
		return nil // nothing to do
	case current == "cancelled":
		// %w "wraps" the error: errors.Is(err, ErrStatusNotAllowed) is still true.
		return fmt.Errorf("%w: cancelled orders cannot be changed", ErrStatusNotAllowed)
	case ownerID != 0 && current != "pending":
		return fmt.Errorf("%w: only pending orders can be cancelled", ErrStatusNotAllowed)
	}

	if newStatus == "cancelled" {
		// Put the items back in stock (one UPDATE ... FROM for all items).
		if _, err := tx.ExecContext(ctx, `
			UPDATE products p
			SET stock = p.stock + oi.quantity, updated_at = NOW()
			FROM order_items oi
			WHERE oi.order_id = $1 AND oi.product_id = p.id`, orderID); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, newStatus, orderID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	if newStatus == "cancelled" {
		h.invalidateProductCache(ctx)
	}
	return nil
}

// respondStatusChange turns the result of changeOrderStatus into an HTTP response.
func (h *Handler) respondStatusChange(w http.ResponseWriter, r *http.Request, orderID int, err error) {
	switch {
	case errors.Is(err, ErrOrderNotFound):
		utils.WriteError(w, http.StatusNotFound, err.Error())
		return
	case errors.Is(err, ErrStatusNotAllowed):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		utils.WriteError(w, http.StatusInternalServerError, "could not update order")
		return
	}
	orders, err := h.fetchOrders(r.Context(), "WHERE o.id = $1", orderID)
	if err != nil || len(orders) == 0 {
		utils.WriteError(w, http.StatusInternalServerError, "could not load order")
		return
	}
	utils.WriteJSON(w, http.StatusOK, orders[0])
}
