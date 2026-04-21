package repository

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderStats holds aggregate order statistics.
type OrderStats struct {
	TotalOrders  int64   `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
	PendingCount int64   `json:"pending_count"`
	ShippedCount int64   `json:"shipped_count"`
}

// SellerStats is the authoritative per-seller dashboard figure set. All counts
// and sums are computed from live DB rows — never derived from product-level
// caches that may drift with price changes.
type SellerStats struct {
	TotalProducts  int64   `json:"total_products"`
	TotalOrders    int64   `json:"total_orders"`     // distinct orders containing this seller's items
	PendingOrders  int64   `json:"pending_orders"`   // pending / confirmed / processing
	UnitsSold      int64   `json:"units_sold"`       // excluding cancelled items
	TotalRevenue   float64 `json:"total_revenue"`    // SUM(price * quantity) for non-cancelled items
	AverageRating  float64 `json:"average_rating"`   // avg of product avg_rating, rated products only
}

// OrderRepository handles all database operations for orders and order items.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository instance.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create inserts a new order and its items within a transaction, also decrementing
// the stock of each product in the order.
func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the order record (including its Items via GORM association).
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("order_repo: failed to create order: %w", err)
		}

		// Decrement stock for each order item.
		for _, item := range order.Items {
			if item.ProductID == nil {
				continue
			}
			result := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", *item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).
				UpdateColumn("total_sales", gorm.Expr("total_sales + ?", item.Quantity))
			if result.Error != nil {
				return fmt.Errorf("order_repo: failed to update stock for product %s: %w", item.ProductID, result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("order_repo: insufficient stock for product %s", item.ProductID)
			}
		}

		return nil
	})
}

// FindByID retrieves an order by its UUID, preloading Items and their Products.
func (r *OrderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := r.db.
		Preload("Items").
		Preload("Items.Product").
		Preload("Buyer").
		First(&order, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("order_repo: order not found: %w", err)
	}
	return &order, nil
}

// ListByBuyer returns a paginated list of orders for a specific buyer.
func (r *OrderRepository) ListByBuyer(buyerID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("buyer_id = ?", buyerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to count buyer orders: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.
		Preload("Items").
		Preload("Items.Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to list buyer orders: %w", err)
	}

	return orders, total, nil
}

// ListBySeller returns a paginated list of orders that contain items from a given seller.
func (r *OrderRepository) ListBySeller(sellerID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	// Find order IDs containing items from this seller.
	subQuery := r.db.Model(&models.OrderItem{}).
		Select("DISTINCT order_id").
		Where("seller_id = ?", sellerID)

	query := r.db.Model(&models.Order{}).Where("id IN (?)", subQuery)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to count seller orders: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.
		Preload("Items", "seller_id = ?", sellerID).
		Preload("Items.Product").
		Preload("Buyer").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to list seller orders: %w", err)
	}

	return orders, total, nil
}

// ListAll returns a paginated list of all orders, optionally filtered by status (for admin use).
func (r *OrderRepository) ListAll(page, perPage int, status string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to count all orders: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.
		Preload("Items").
		Preload("Items.Product").
		Preload("Buyer").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("order_repo: failed to list all orders: %w", err)
	}

	return orders, total, nil
}

// UpdateStatus updates the status of an order.
func (r *OrderRepository) UpdateStatus(id uuid.UUID, status string) error {
	result := r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("order_repo: failed to update order status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order_repo: order not found")
	}
	return nil
}

// CancelAndRestoreStock transitions an order to 'cancelled' and adds its items'
// quantities back to the product stock. Wrapped in a transaction so stock and
// order state can't drift. Idempotent: if the order is already cancelled we
// exit early without touching stock again (guards against duplicate webhook
// failure callbacks releasing stock twice).
func (r *OrderRepository) CancelAndRestoreStock(orderID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Preload("Items").First(&order, "id = ?", orderID).Error; err != nil {
			return fmt.Errorf("order_repo: cancel find: %w", err)
		}
		if order.Status == "cancelled" {
			return nil
		}
		// Don't release stock for an order that was actually paid — that would
		// corrupt inventory. Callers should never invoke this for paid orders,
		// but defending in-depth is cheap.
		if order.PaymentStatus == "paid" {
			return fmt.Errorf("order_repo: cannot cancel a paid order")
		}

		if err := tx.Model(&models.Order{}).
			Where("id = ?", orderID).
			Update("status", "cancelled").Error; err != nil {
			return fmt.Errorf("order_repo: cancel update: %w", err)
		}

		for _, item := range order.Items {
			if item.ProductID == nil {
				continue
			}
			if err := tx.Model(&models.Product{}).
				Where("id = ?", *item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).
				UpdateColumn("total_sales", gorm.Expr("GREATEST(0, total_sales - ?)", item.Quantity)).
				Error; err != nil {
				return fmt.Errorf("order_repo: restore stock for %s: %w", item.ProductID, err)
			}
		}
		return nil
	})
}

// MarkPaid transitions an order to payment_status='paid' and status='confirmed',
// storing the payment provider's reference. Uses an optimistic guard so a
// duplicate success callback can't flip a later 'cancelled' or 'refunded'
// order back to 'confirmed'.
func (r *OrderRepository) MarkPaid(orderID uuid.UUID, providerRef string) error {
	result := r.db.Model(&models.Order{}).
		Where("id = ? AND payment_status <> ?", orderID, "paid").
		Updates(map[string]any{
			"payment_status":    "paid",
			"status":            "confirmed",
			"payment_intent_id": providerRef,
		})
	if result.Error != nil {
		return fmt.Errorf("order_repo: mark paid: %w", result.Error)
	}
	return nil
}

// GetStats returns aggregate order statistics: total orders, revenue, and counts by status.
func (r *OrderRepository) GetStats() (*OrderStats, error) {
	var stats OrderStats

	if err := r.db.Model(&models.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("order_repo: failed to get total orders: %w", err)
	}

	if err := r.db.Model(&models.Order{}).
		Select("COALESCE(SUM(total), 0)").
		Where("payment_status = ?", "paid").
		Row().Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("order_repo: failed to get total revenue: %w", err)
	}

	if err := r.db.Model(&models.Order{}).Where("status = ?", "pending").Count(&stats.PendingCount).Error; err != nil {
		return nil, fmt.Errorf("order_repo: failed to get pending count: %w", err)
	}

	if err := r.db.Model(&models.Order{}).Where("status = ?", "shipped").Count(&stats.ShippedCount).Error; err != nil {
		return nil, fmt.Errorf("order_repo: failed to get shipped count: %w", err)
	}

	return &stats, nil
}

// GetSellerStats aggregates live per-seller figures directly from order_items
// and products. Uses historical order_items.price (not product.price) so
// price changes don't corrupt past revenue.
func (r *OrderRepository) GetSellerStats(sellerID uuid.UUID) (*SellerStats, error) {
	var stats SellerStats

	if err := r.db.Model(&models.Product{}).
		Where("seller_id = ?", sellerID).
		Count(&stats.TotalProducts).Error; err != nil {
		return nil, fmt.Errorf("order_repo: seller product count: %w", err)
	}

	// Distinct orders containing this seller's items (any status).
	if err := r.db.Model(&models.OrderItem{}).
		Where("seller_id = ?", sellerID).
		Distinct("order_id").
		Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("order_repo: seller total orders: %w", err)
	}

	// Orders currently in flight (seller has work to do).
	if err := r.db.Table("order_items AS oi").
		Joins("JOIN orders o ON o.id = oi.order_id").
		Where("oi.seller_id = ? AND o.status IN ?", sellerID,
			[]string{"pending", "confirmed", "processing"}).
		Distinct("oi.order_id").
		Count(&stats.PendingOrders).Error; err != nil {
		return nil, fmt.Errorf("order_repo: seller pending orders: %w", err)
	}

	// Units sold + revenue, excluding items tied to cancelled orders.
	var agg struct {
		Units   int64
		Revenue float64
	}
	if err := r.db.Table("order_items AS oi").
		Joins("JOIN orders o ON o.id = oi.order_id").
		Where("oi.seller_id = ? AND o.status <> ?", sellerID, "cancelled").
		Select("COALESCE(SUM(oi.quantity), 0) AS units, COALESCE(SUM(oi.price * oi.quantity), 0) AS revenue").
		Row().Scan(&agg.Units, &agg.Revenue); err != nil {
		return nil, fmt.Errorf("order_repo: seller revenue: %w", err)
	}
	stats.UnitsSold = agg.Units
	stats.TotalRevenue = agg.Revenue

	// Average rating across this seller's rated products only.
	if err := r.db.Model(&models.Product{}).
		Where("seller_id = ? AND avg_rating > 0", sellerID).
		Select("COALESCE(AVG(avg_rating), 0)").
		Row().Scan(&stats.AverageRating); err != nil {
		return nil, fmt.Errorf("order_repo: seller avg rating: %w", err)
	}

	return &stats, nil
}

// GenerateOrderNumber creates a unique order number in the format ART-XXXXXX.
func (r *OrderRepository) GenerateOrderNumber() string {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[src.Intn(len(chars))]
	}
	orderNum := fmt.Sprintf("ART-%s", string(code))

	// Ensure uniqueness by checking the database.
	var count int64
	r.db.Model(&models.Order{}).Where("order_number = ?", orderNum).Count(&count)
	if count > 0 {
		return r.GenerateOrderNumber() // Retry on collision.
	}

	return orderNum
}
