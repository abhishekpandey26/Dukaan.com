// Package middleware contains functions that wrap handlers to add behaviour
// (logging, CORS, panic recovery, authentication).
//
// A middleware is just a function that takes a handler and returns a new handler:
//
//	func(next http.Handler) http.Handler
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"gocommerce/auth"
	"gocommerce/utils"
)

// ---------- Logging ----------

// statusRecorder "embeds" http.ResponseWriter, so it has all its methods,
// but we override WriteHeader to remember the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// ---------- Recover ----------

// Recover catches panics so one bad request doesn't crash the whole server.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				utils.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ---------- CORS ----------

// CORS lets the React app (on another port) call this API from the browser.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions { // browser "preflight" request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------- Auth ----------

// We use our own type for context keys so they can't clash with other packages.
type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

// UserID reads the logged-in user's id that Auth stored in the request context.
func UserID(ctx context.Context) int {
	id, _ := ctx.Value(userIDKey).(int) // type assertion; "_" ignores the ok flag
	return id
}

func Role(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}

// Auth only lets the request through if it has a valid "Authorization: Bearer <token>" header.
func Auth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			utils.WriteError(w, http.StatusUnauthorized, "missing or invalid Authorization header")
			return
		}

		claims, err := auth.ParseToken(secret, token)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Attach user info to the request so handlers can read it.
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, roleKey, claims.Role)
		next(w, r.WithContext(ctx))
	}
}

// Admin = Auth + "role must be admin".
func Admin(secret string, next http.HandlerFunc) http.HandlerFunc {
	return Auth(secret, func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != "admin" {
			utils.WriteError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	})
}
