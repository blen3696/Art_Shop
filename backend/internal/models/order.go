package models

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a purchase order placed by a buyer.
type Order struct {
	ID                    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	BuyerID               uuid.UUID `gorm:"type:uuid;not null;index" json:"buyer_id"`
	OrderNumber           string    `gorm:"uniqueIndex;not null;size:20" json:"order_number"`
	Status                string    `gorm:"size:30;default:pending" json:"status"`
	Subtotal              float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	ShippingCost          float64   `gorm:"type:decimal(10,2);default:0" json:"shipping_cost"`
	Tax                   float64   `gorm:"type:decimal(10,2);default:0" json:"tax"`
	Discount              float64   `gorm:"type:decimal(10,2);default:0" json:"discount"`
	Total                 float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	ShippingName          *string   `gorm:"size:255" json:"shipping_name"`
	ShippingAddressLine1  *string   `gorm:"size:255" json:"shipping_address_line1"`
	ShippingAddressLine2  *string   `gorm:"size:255" json:"shipping_address_line2"`
	ShippingCity          *string   `gorm:"size:100" json:"shipping_city"`
	ShippingState         *string   `gorm:"size:100" json:"shipping_state"`
	ShippingCountry       *string   `gorm:"size:100" json:"shipping_country"`
	ShippingZip           *string   `gorm:"size:20" json:"shipping_zip"`
	ShippingPhone         *string   `gorm:"size:20" json:"shipping_phone"`
	PaymentMethod         *string   `gorm:"size:50" json:"payment_method"`
	PaymentStatus         string    `gorm:"size:30;default:pending" json:"payment_status"`
	PaymentIntentID       *string   `gorm:"size:255" json:"payment_intent_id"`
	TrackingNumber        *string   `gorm:"size:100" json:"tracking_number"`
	Notes                 *string   `json:"notes"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Relations
	Buyer User        `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// TableName overrides the default table name.
func (Order) TableName() string {
	return "orders"
}

// OrderItem represents an individual line item within an order.
type OrderItem struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrderID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID *uuid.UUID `gorm:"type:uuid" json:"product_id"`
	SellerID  *uuid.UUID `gorm:"type:uuid;index" json:"seller_id"`
	Title     string     `gorm:"not null;size:255" json:"title"`
	Price     float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity  int        `gorm:"not null" json:"quantity"`
	Thumbnail *string    `json:"thumbnail"`
	Status    string     `gorm:"size:30;default:pending" json:"status"`
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	Order   Order    `gorm:"foreignKey:OrderID" json:"-"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Seller  *User    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}

// TableName overrides the default table name.
func (OrderItem) TableName() string {
	return "order_items"
}

// CreateOrderRequest is the payload for placing a new order.
type CreateOrderRequest struct {
	ShippingName         string `json:"shipping_name" validate:"required"`
	ShippingAddressLine1 string `json:"shipping_address_line1" validate:"required"`
	ShippingAddressLine2 string `json:"shipping_address_line2"`
	ShippingCity         string `json:"shipping_city" validate:"required"`
	ShippingState        string `json:"shipping_state" validate:"required"`
	ShippingCountry      string `json:"shipping_country" validate:"required"`
	ShippingZip          string `json:"shipping_zip" validate:"required"`
	ShippingPhone        string `json:"shipping_phone"`
	PaymentMethod        string `json:"payment_method" validate:"required"`
	Notes                string `json:"notes"`
}

// UpdateOrderStatusRequest is used by admin/seller to change order status.
type UpdateOrderStatusRequest struct {
	Status         string `json:"status" validate:"required"`
	TrackingNumber string `json:"tracking_number"`
}

// OrderListQuery holds query parameters for listing orders.
type OrderListQuery struct {
	Page          int    `json:"page"`
	PerPage       int    `json:"per_page"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	SortBy        string `json:"sort_by"`    // created_at, total
	SortOrder     string `json:"sort_order"` // asc, desc
}

// OrderStatusValues enumerates the valid order statuses.
var OrderStatusValues = []string{
	"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded",
}

// PaymentStatusValues enumerates the valid payment statuses.
var PaymentStatusValues = []string{
	"pending", "paid", "failed", "refunded",
}

// OrderItemStatusValues enumerates the valid order item statuses.
var OrderItemStatusValues = []string{
	"pending", "confirmed", "shipped", "delivered", "cancelled",
}
