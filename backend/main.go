// GoCommerce API - entry point.
//
// Startup order: config -> postgres -> tables -> seed -> redis -> HTTP server.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocommerce/cache"
	"gocommerce/config"
	"gocommerce/db"
	"gocommerce/handlers"
	"gocommerce/middleware"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := db.Seed(database, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	redisCache := cache.New(cfg.RedisAddr)
	defer redisCache.Close()

	h := handlers.New(database, redisCache, cfg.JWTSecret)

	// Middlewares wrap each other like layers of an onion:
	// Logger -> Recover -> CORS -> router -> (Auth) -> handler
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.Logger(middleware.Recover(middleware.CORS(routes(h, cfg.JWTSecret)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run the server in a goroutine so main can wait for a shutdown signal.
	go func() {
		log.Printf("GoCommerce API listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown: wait for Ctrl+C (SIGINT) or `docker stop` / Kubernetes (SIGTERM),
	// then give in-flight requests up to 10 seconds to finish.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // blocks here until a signal arrives

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("bye")
}
