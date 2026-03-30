package repository

import (
	"fmt"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewRepository handles all database operations for product reviews.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository creates a new ReviewRepository instance.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create inserts a new review record into the database.
func (r *ReviewRepository) Create(review *models.Review) error {
	if err := r.db.Create(review).Error; err != nil {
		return fmt.Errorf("review_repo: failed to create review: %w", err)
	}
	return nil
}

// GetByProduct returns a paginated list of reviews for a specific product, preloading the User.
func (r *ReviewRepository) GetByProduct(productID uuid.UUID, page, perPage int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	query := r.db.Model(&models.Review{}).Where("product_id = ?", productID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("review_repo: failed to count reviews: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("review_repo: failed to list reviews: %w", err)
	}

	return reviews, total, nil
}

// GetByUser returns all reviews written by a specific user.
func (r *ReviewRepository) GetByUser(userID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.
		Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("review_repo: failed to list user reviews: %w", err)
	}
	return reviews, nil
}

// Delete removes a review only if it belongs to the specified user.
func (r *ReviewRepository) Delete(id, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Review{})
	if result.Error != nil {
		return fmt.Errorf("review_repo: failed to delete review: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("review_repo: review not found or not owned by user")
	}
	return nil
}

// HasPurchased checks whether the user has purchased the specified product
// (i.e., has a delivered order containing that product).
func (r *ReviewRepository) HasPurchased(userID, productID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.buyer_id = ? AND order_items.product_id = ? AND orders.status IN ?",
			userID, productID, []string{"delivered", "confirmed", "processing", "shipped"}).
		Count(&count)
	return count > 0
}
