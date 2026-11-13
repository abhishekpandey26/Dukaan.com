package main

import (
	"net/http"

	"gocommerce/handlers"
	"gocommerce/middleware"
)

// routes registers every endpoint in one place, so you can see the whole API at a glance.
//
// Since Go 1.22 the standard http.ServeMux understands methods and path
// parameters: "GET /api/products/{id}" -> read it with r.PathValue("id").
// No external router (Gin, chi...) needed!
func routes(h *handlers.Handler, jwtSecret string) *http.ServeMux {
	mux := http.NewServeMux()

	// Small helpers so the lines below stay short.
	auth := func(f http.HandlerFunc) http.HandlerFunc { return middleware.Auth(jwtSecret, f) }
	admin := func(f http.HandlerFunc) http.HandlerFunc { return middleware.Admin(jwtSecret, f) }

	// Public
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/products", h.ListProducts)
	mux.HandleFunc("GET /api/products/{id}", h.GetProduct)
	mux.HandleFunc("GET /api/categories", h.ListCategories)

	// Logged-in users
	mux.HandleFunc("GET /api/auth/me", auth(h.Me))

	mux.HandleFunc("GET /api/cart", auth(h.GetCart))
	mux.HandleFunc("POST /api/cart", auth(h.AddToCart))
	mux.HandleFunc("PUT /api/cart/{productId}", auth(h.UpdateCartItem))
	mux.HandleFunc("DELETE /api/cart/{productId}", auth(h.RemoveFromCart))
	mux.HandleFunc("DELETE /api/cart", auth(h.ClearCart))

	mux.HandleFunc("POST /api/orders", auth(h.PlaceOrder))
	mux.HandleFunc("GET /api/orders", auth(h.ListMyOrders))
	mux.HandleFunc("GET /api/orders/{id}", auth(h.GetMyOrder))
	mux.HandleFunc("POST /api/orders/{id}/cancel", auth(h.CancelMyOrder))

	// Admin only
	mux.HandleFunc("GET /api/admin/stats", admin(h.Stats))
	mux.HandleFunc("POST /api/admin/products", admin(h.CreateProduct))
	mux.HandleFunc("PUT /api/admin/products/{id}", admin(h.UpdateProduct))
	mux.HandleFunc("DELETE /api/admin/products/{id}", admin(h.DeleteProduct))
	mux.HandleFunc("GET /api/admin/orders", admin(h.ListAllOrders))
	mux.HandleFunc("PUT /api/admin/orders/{id}/status", admin(h.UpdateOrderStatus))
	mux.HandleFunc("GET /api/admin/users", admin(h.ListUsers))

	return mux
}
