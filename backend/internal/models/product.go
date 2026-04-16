package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Category represents a hierarchical product category.
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string     `gorm:"not null;size:100" json:"name"`
	Slug        string     `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	Description *string    `json:"description"`
	ImageURL    *string    `json:"image_url"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relations
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName overrides the default table name.
func (Category) TableName() string {
	return "categories"
}

// Product represents an art product listed by a seller.
type Product struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SellerID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"seller_id"`
	Title           string         `gorm:"not null;size:255" json:"title"`
	Description     *string        `json:"description"`
	Price           float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	CompareAtPrice  *float64       `gorm:"type:decimal(10,2)" json:"compare_at_price"`
	CategoryID      *uuid.UUID     `gorm:"type:uuid;index" json:"category_id"`
	Images          pq.StringArray `gorm:"type:text[];default:'{}'" json:"images"`
	Thumbnail       *string        `json:"thumbnail"`
	Stock           int            `gorm:"default:0" json:"stock"`
	SKU             *string        `gorm:"size:100" json:"sku"`
	Tags            pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	Medium          *string        `gorm:"size:100" json:"medium"`
	Dimensions      *string        `gorm:"size:100" json:"dimensions"`
	Weight          *float64       `gorm:"type:decimal(8,2)" json:"weight"`
	IsPublished     bool           `gorm:"default:true;index" json:"is_published"`
	IsFeatured      bool           `gorm:"default:false;index" json:"is_featured"`
	AvgRating       float64        `gorm:"type:decimal(3,2);default:0" json:"avg_rating"`
	TotalReviews    int            `gorm:"default:0" json:"total_reviews"`
	TotalSales      int            `gorm:"default:0" json:"total_sales"`
	ViewCount       int            `gorm:"default:0" json:"view_count"`
	AIDescription   *string        `json:"ai_description"`
	AITags          pq.StringArray `gorm:"type:text[];default:'{}'" json:"ai_tags"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relations
	Seller   User      `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName overrides the default table name.
func (Product) TableName() string {
	return "products"
}

// CreateProductRequest is the payload for creating a new product.
type CreateProductRequest struct {
	Title          string   `json:"title" validate:"required,min=2"`
	Description    *string  `json:"description"`
	Price          float64  `json:"price" validate:"required,gte=0"`
	CompareAtPrice *float64 `json:"compare_at_price"`
	CategoryID     *string  `json:"category_id"`
	Images         []string `json:"images"`
	Thumbnail      *string  `json:"thumbnail"`
	Stock          int      `json:"stock" validate:"gte=0"`
	SKU            *string  `json:"sku"`
	Tags           []string `json:"tags"`
	Medium         *string  `json:"medium"`
	Dimensions     *string  `json:"dimensions"`
	Weight         *float64 `json:"weight"`
	IsPublished    *bool    `json:"is_published"`
}

// UpdateProductRequest is the payload for updating a product.
type UpdateProductRequest struct {
	Title          *string  `json:"title"`
	Description    *string  `json:"description"`
	Price          *float64 `json:"price"`
	CompareAtPrice *float64 `json:"compare_at_price"`
	CategoryID     *string  `json:"category_id"`
	Images         []string `json:"images"`
	Thumbnail      *string  `json:"thumbnail"`
	Stock          *int     `json:"stock"`
	SKU            *string  `json:"sku"`
	Tags           []string `json:"tags"`
	Medium         *string  `json:"medium"`
	Dimensions     *string  `json:"dimensions"`
	Weight         *float64 `json:"weight"`
	IsPublished    *bool    `json:"is_published"`
	IsFeatured     *bool    `json:"is_featured"`
}

// ProductListQuery holds query parameters for listing/filtering products.
type ProductListQuery struct {
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	Search     string  `json:"search"`
	CategoryID string  `json:"category_id"`
	SellerID   string  `json:"seller_id"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	Medium     string  `json:"medium"`
	SortBy     string  `json:"sort_by"`   // price, created_at, avg_rating, total_sales
	SortOrder  string  `json:"sort_order"` // asc, desc
	IsFeatured *bool   `json:"is_featured"`
	Tags       []string `json:"tags"`
}
