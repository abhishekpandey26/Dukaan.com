package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gocommerce/models"
	"gocommerce/utils"
)

// All product cache keys start with "gocommerce:products:" so we can clear them in one go.
const productCacheTTL = 5 * time.Minute

const productColumns = `id, name, description, price_cents, stock, category, image_url, created_at, updated_at`

// scanner is an interface satisfied by both *sql.Row and *sql.Rows.
// Because of it, scanProduct works for single rows AND for loops over many rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(s scanner) (models.Product, error) {
	var p models.Product
	err := s.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Stock,
		&p.Category, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// GET /api/products?search=shoe&category=Fashion&page=1&limit=12
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	page := max(queryInt(r, "page", 1), 1)
	limit := min(max(queryInt(r, "limit", 12), 1), 100)

	// 1. Try the cache first.
	cacheKey := fmt.Sprintf("gocommerce:products:list:search=%s:category=%s:page=%d:limit=%d", search, category, page, limit)
	var cached models.ProductList
	if h.Cache.Get(ctx, cacheKey, &cached) {
		w.Header().Set("X-Cache", "HIT") // look at this header in the browser dev tools!
		utils.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// 2. Build the WHERE clause step by step.
	// We NEVER put user input directly into SQL; we use $1, $2... placeholders
	// and pass values in args. This prevents SQL injection.
	var conditions []string
	var args []any
	if search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", len(args), len(args)))
	}
	if category != "" {
		args = append(args, category)
		conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM products "+where, args...).Scan(&total); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not count products")
		return
	}

	args = append(args, limit, (page-1)*limit)
	query := fmt.Sprintf("SELECT %s FROM products %s ORDER BY id DESC LIMIT $%d OFFSET $%d",
		productColumns, where, len(args)-1, len(args))

	rows, err := h.DB.QueryContext(ctx, query, args...)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load products")
		return
	}
	defer rows.Close() // always close rows, or the DB connection leaks

	products := []models.Product{} // empty slice (not nil) so JSON gives [] instead of null
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "could not read product")
			return
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load products")
		return
	}

	// 3. Save to cache for next time.
	result := models.ProductList{Products: products, Total: total, Page: page, Limit: limit}
	h.Cache.Set(ctx, cacheKey, result, productCacheTTL)
	w.Header().Set("X-Cache", "MISS")
	utils.WriteJSON(w, http.StatusOK, result)
}

// GET /api/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r, "id")
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	cacheKey := fmt.Sprintf("gocommerce:products:item:%d", id)
	var product models.Product
	if h.Cache.Get(r.Context(), cacheKey, &product) {
		w.Header().Set("X-Cache", "HIT")
		utils.WriteJSON(w, http.StatusOK, product)
		return
	}

	row := h.DB.QueryRowContext(r.Context(), "SELECT "+productColumns+" FROM products WHERE id = $1", id)
	product, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load product")
		return
	}

	h.Cache.Set(r.Context(), cacheKey, product, productCacheTTL)
	w.Header().Set("X-Cache", "MISS")
	utils.WriteJSON(w, http.StatusOK, product)
}

// GET /api/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	const cacheKey = "gocommerce:products:categories"
	categories := []string{}
	if h.Cache.Get(r.Context(), cacheKey, &categories) {
		utils.WriteJSON(w, http.StatusOK, categories)
		return
	}

	rows, err := h.DB.QueryContext(r.Context(),
		`SELECT DISTINCT category FROM products WHERE category <> '' ORDER BY category`)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "could not load categories")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err == nil {
			categories = append(categories, c)
		}
	}

	h.Cache.Set(r.Context(), cacheKey, categories, productCacheTTL)
	utils.WriteJSON(w, http.StatusOK, categories)
}

// invalidateProductCache is called whenever products change (admin edit, order placed...).
// Simple rule: when in doubt, clear everything product-related.
func (h *Handler) invalidateProductCache(ctx context.Context) {
	h.Cache.DeletePattern(ctx, "gocommerce:products:*")
}
