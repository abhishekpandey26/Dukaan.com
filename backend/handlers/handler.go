// Package handlers contains the HTTP handlers (the "controllers") of the API.
package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"gocommerce/cache"
	"gocommerce/utils"
)

// Handler holds the dependencies every handler needs.
// Passing them in a struct (instead of global variables) is called
// "dependency injection" and makes the code easy to test.
type Handler struct {
	DB        *sql.DB
	Cache     *cache.Cache
	JWTSecret string
}

func New(db *sql.DB, c *cache.Cache, jwtSecret string) *Handler {
	return &Handler{DB: db, Cache: c, JWTSecret: jwtSecret}
}

// Health is used by Docker/Kubernetes to check the app is alive.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.DB.PingContext(ctx); err != nil {
		utils.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// pathInt reads a path parameter like {id} from "GET /api/products/{id}" as an int.
func pathInt(r *http.Request, name string) (int, bool) {
	n, err := strconv.Atoi(r.PathValue(name))
	return n, err == nil && n > 0
}

// queryInt reads ?name=123 from the URL, or returns fallback.
func queryInt(r *http.Request, name string, fallback int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		return fallback
	}
	return n
}
