package repository

import (
	"fmt"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CartRepository handles all database operations for carts and wishlists.
type CartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a new CartRepository instance.
func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// GetByUser returns all cart items for a user, preloading the associated Product.
func (r *CartRepository) GetByUser(userID uuid.UUID) ([]models.CartItem, error) {
	var items []models.CartItem
	if err := r.db.
		Where("user_id = ?", userID).
		Preload("Product").
		Preload("Product.Seller").
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("cart_repo: failed to get cart items: %w", err)
	}
	return items, nil
}

// AddItem performs an upsert: if the user already has this product in their cart,
// the quantity is incremented; otherwise a new cart item is created.
func (r *CartRepository) AddItem(item *models.CartItem) error {
	// Use ON CONFLICT to upsert: if (user_id, product_id) exists, increment quantity.
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"quantity": gorm.Expr("cart_items.quantity + ?", item.Quantity)}),
	}).Create(item)

	if result.Error != nil {
		return fmt.Errorf("cart_repo: failed to add item to cart: %w", result.Error)
	}
	return nil
}

// UpdateQuantity sets the quantity for a specific cart item. If quantity is 0 or less, the item is removed.
func (r *CartRepository) UpdateQuantity(userID, productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return r.RemoveItem(userID, productID)
	}

	result := r.db.Model(&models.CartItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Update("quantity", quantity)

	if result.Error != nil {
		return fmt.Errorf("cart_repo: failed to update cart item quantity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("cart_repo: cart item not found")
	}
	return nil
}

// RemoveItem deletes a specific item from the user's cart.
func (r *CartRepository) RemoveItem(userID, productID uuid.UUID) error {
	result := r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.CartItem{})
	if result.Error != nil {
		return fmt.Errorf("cart_repo: failed to remove cart item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("cart_repo: cart item not found")
	}
	return nil
}

// Clear removes all items from a user's cart.
func (r *CartRepository) Clear(userID uuid.UUID) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		return fmt.Errorf("cart_repo: failed to clear cart: %w", err)
	}
	return nil
}

// GetWishlist returns all wishlist items for a user, preloading the associated Product.
func (r *CartRepository) GetWishlist(userID uuid.UUID) ([]models.Wishlist, error) {
	var items []models.Wishlist
	if err := r.db.
		Where("user_id = ?", userID).
		Preload("Product").
		Preload("Product.Seller").
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("cart_repo: failed to get wishlist: %w", err)
	}
	return items, nil
}

// AddToWishlist adds a product to the user's wishlist. If it already exists, no error is returned.
func (r *CartRepository) AddToWishlist(item *models.Wishlist) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoNothing: true,
	}).Create(item)

	if result.Error != nil {
		return fmt.Errorf("cart_repo: failed to add to wishlist: %w", result.Error)
	}
	return nil
}

// RemoveFromWishlist removes a product from the user's wishlist.
func (r *CartRepository) RemoveFromWishlist(userID, productID uuid.UUID) error {
	result := r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Wishlist{})
	if result.Error != nil {
		return fmt.Errorf("cart_repo: failed to remove from wishlist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("cart_repo: wishlist item not found")
	}
	return nil
}
