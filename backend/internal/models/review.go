package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Review represents a product review and rating from a buyer.
type Review struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProductID          uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_review_product_user" json:"product_id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_review_product_user" json:"user_id"`
	OrderID            *uuid.UUID     `gorm:"type:uuid" json:"order_id"`
	Rating             int            `gorm:"not null" json:"rating"`
	Title              *string        `gorm:"size:255" json:"title"`
	Comment            *string        `json:"comment"`
	Images             pq.StringArray `gorm:"type:text[];default:'{}'" json:"images"`
	IsVerifiedPurchase bool           `gorm:"default:false" json:"is_verified_purchase"`
	HelpfulCount       int            `gorm:"default:0" json:"helpful_count"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	// Relations (for preloading)
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Order   *Order  `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName overrides the default table name.
func (Review) TableName() string {
	return "reviews"
}

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// CreateReviewRequest is the payload for submitting a review.
type CreateReviewRequest struct {
	ProductID string   `json:"product_id" validate:"required,uuid"`
	OrderID   *string  `json:"order_id"`
	Rating    int      `json:"rating" validate:"required,min=1,max=5"`
	Title     *string  `json:"title"`
	Comment   *string  `json:"comment"`
	Images    []string `json:"images"`
}

// UpdateReviewRequest allows a user to edit their review.
type UpdateReviewRequest struct {
	Rating  *int     `json:"rating" validate:"omitempty,min=1,max=5"`
	Title   *string  `json:"title"`
	Comment *string  `json:"comment"`
	Images  []string `json:"images"`
}

// ReviewListQuery holds query parameters for listing reviews.
type ReviewListQuery struct {
	Page      int    `json:"page"`
	PerPage   int    `json:"per_page"`
	ProductID string `json:"product_id"`
	UserID    string `json:"user_id"`
	MinRating int    `json:"min_rating"`
	SortBy    string `json:"sort_by"`    // created_at, rating, helpful_count
	SortOrder string `json:"sort_order"` // asc, desc
}
