package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductReviewSummary is a cached LLM-generated summary of a product's reviews.
type ProductReviewSummary struct {
	ProductID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"product_id"`
	Summary     string    `gorm:"not null" json:"summary"`
	ReviewCount int       `gorm:"not null" json:"review_count"`
	GeneratedAt time.Time `json:"generated_at"`
}

func (ProductReviewSummary) TableName() string {
	return "product_review_summaries"
}
