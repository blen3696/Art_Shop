package models

import (
	"time"

	"github.com/google/uuid"
)

// CartItem represents a product in a user's shopping cart.
type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_cart_user_product" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cart_user_product" json:"product_id"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`

	// Relations (for preloading)
	User    User    `gorm:"foreignKey:UserID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName overrides the default table name.
func (CartItem) TableName() string {
	return "cart_items"
}

// Wishlist represents a product saved to a user's wishlist.
type Wishlist struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_wishlist_user_product" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_wishlist_user_product" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relations (for preloading)
	User    User    `gorm:"foreignKey:UserID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName overrides the default table name.
func (Wishlist) TableName() string {
	return "wishlists"
}

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// AddToCartRequest is the payload for adding a product to the cart.
type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateCartItemRequest is the payload for changing the quantity of a cart item.
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

// AddToWishlistRequest is the payload for adding a product to the wishlist.
type AddToWishlistRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
}

// CartResponse provides a summary of the cart contents.
type CartResponse struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	Subtotal   float64    `json:"subtotal"`
}
