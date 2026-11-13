// Package db opens the PostgreSQL connection, creates tables and seeds demo data.
package db

import (
	"database/sql"
	_ "embed" // needed for the //go:embed directive below
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // registers the "postgres" driver with database/sql

	"gocommerce/auth"
)

//go:embed schema.sql
var schema string

// Connect opens a connection pool and waits until Postgres is ready.
// The retry loop matters in Docker/Kubernetes, where the DB may start after us.
func Connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url) // does NOT connect yet, just prepares the pool
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	for attempt := 1; attempt <= 10; attempt++ {
		if err = db.Ping(); err == nil {
			log.Println("connected to postgres")
			return db, nil
		}
		log.Printf("postgres not ready (attempt %d/10): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to postgres: %w", err)
}

// Migrate creates the tables if they don't exist yet.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

// Seed creates the admin user and a few demo products (only if the table is empty).
func Seed(db *sql.DB, adminEmail, adminPassword string) error {
	hash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return err
	}
	// ON CONFLICT DO NOTHING = "insert unless this email already exists"
	_, err = db.Exec(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ('Admin', $1, $2, 'admin')
		ON CONFLICT (email) DO NOTHING`, adminEmail, hash)
	if err != nil {
		return err
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Images live in frontend/public/images, so they work without internet.
	demo := []struct {
		name, desc, category, image string
		price                       int64
		stock                       int
	}{
		{"Wireless Headphones", "Noise cancelling over-ear headphones with 30h battery.", "Electronics", "/images/headphones.svg", 7999, 25},
		{"Mechanical Keyboard", "Hot-swappable keyboard with RGB lighting.", "Electronics", "/images/keyboard.svg", 5999, 15},
		{"Running Shoes", "Lightweight shoes for daily running.", "Fashion", "/images/shoes.svg", 4499, 40},
		{"Cotton T-Shirt", "Soft 100% cotton crew neck tee.", "Fashion", "/images/tshirt.svg", 1299, 100},
		{"Coffee Mug", "Ceramic mug, 350ml, dishwasher safe.", "Home", "/images/mug.svg", 899, 60},
		{"Desk Lamp", "LED desk lamp with adjustable brightness.", "Home", "/images/lamp.svg", 2499, 3},
		{"The Go Programming Language", "Classic book for learning Go.", "Books", "/images/book.svg", 3499, 20},
		{"Water Bottle", "Insulated steel bottle, keeps drinks cold 24h.", "Sports", "/images/bottle.svg", 1999, 0},
	}
	for _, p := range demo {
		_, err := db.Exec(`
			INSERT INTO products (name, description, price_cents, stock, category, image_url)
			VALUES ($1, $2, $3, $4, $5, $6)`, p.name, p.desc, p.price, p.stock, p.category, p.image)
		if err != nil {
			return err
		}
	}
	log.Printf("seeded %d demo products", len(demo))
	return nil
}
