// Package models holds the structs that we send/receive as JSON.
// The `json:"..."` struct tags control the JSON field names.
package models

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means: never put this in JSON
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductList is a page of products plus info for pagination.
type ProductList struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// ProductInput is what the admin sends to create/update a product.
type ProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int    `json:"stock"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

// Validate is a method on ProductInput. It returns nil when the input is OK.
func (p *ProductInput) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Category = strings.TrimSpace(p.Category)
	switch {
	case p.Name == "":
		return errors.New("name is required")
	case p.PriceCents < 0:
		return errors.New("price cannot be negative")
	case p.Stock < 0:
		return errors.New("stock cannot be negative")
	}
	return nil
}

type CartItem struct {
	ProductID     int    `json:"product_id"`
	Name          string `json:"name"`
	PriceCents    int64  `json:"price_cents"`
	ImageURL      string `json:"image_url"`
	Stock         int    `json:"stock"`
	Quantity      int    `json:"quantity"`
	SubtotalCents int64  `json:"subtotal_cents"`
}

type Cart struct {
	Items      []CartItem `json:"items"`
	TotalCents int64      `json:"total_cents"`
}

type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	UserEmail  string      `json:"user_email"`
	TotalCents int64       `json:"total_cents"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID   *int   `json:"product_id"` // pointer because it becomes NULL if the product is deleted
	ProductName string `json:"product_name"`
	PriceCents  int64  `json:"price_cents"`
	Quantity    int    `json:"quantity"`
}

// Order statuses. A map[string]bool is a simple "set" in Go.
var ValidOrderStatuses = map[string]bool{
	"pending":   true,
	"paid":      true,
	"shipped":   true,
	"delivered": true,
	"cancelled": true,
}
