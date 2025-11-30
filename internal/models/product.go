package models

import (
	"time"
)

type Category string

const (
	CategorySmartphone Category = "smartphone"
	CategoryLaptop     Category = "laptop"
	CategoryTablet     Category = "tablet"
	CategoryTV         Category = "tv"
	CategoryCamera     Category = "camera"
)

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    Category  `json:"category"`
	Brand       string    `json:"brand"`
	Model       string    `json:"model"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Category    Category `json:"category" validate:"required"`
	Brand       string   `json:"brand" validate:"required"`
	Model       string   `json:"model" validate:"required"`
	Price       float64  `json:"price" validate:"required,min=0"`
	Stock       int      `json:"stock" validate:"min=0"`
	ImageURL    string   `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Category    *Category `json:"category"`
	Brand       *string   `json:"brand"`
	Model       *string   `json:"model"`
	Price       *float64  `json:"price"`
	Stock       *int      `json:"stock"`
	ImageURL    *string   `json:"image_url"`
}

type ProductFilter struct {
	Category Category `json:"category"`
	Brand    string   `json:"brand"`
	MinPrice float64  `json:"min_price"`
	MaxPrice float64  `json:"max_price"`
	Limit    int      `json:"limit" validate:"min=1,max=100"`
	Offset   int      `json:"offset" validate:"min=0"`
}
