package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/google/uuid"
)

// CreateOrderRequest holds the data required to create an order from the cart.
type CreateOrderRequest struct {
	ShippingName         string `json:"shipping_name"`
	ShippingAddressLine1 string `json:"shipping_address_line1"`
	ShippingAddressLine2 string `json:"shipping_address_line2"`
	ShippingCity         string `json:"shipping_city"`
	ShippingState        string `json:"shipping_state"`
	ShippingCountry      string `json:"shipping_country"`
	ShippingZip          string `json:"shipping_zip"`
	ShippingPhone        string `json:"shipping_phone"`
	PaymentMethod        string `json:"payment_method"`
	Notes                string `json:"notes"`
}

// validStatusTransitions defines the allowed state machine transitions for order status.
var validStatusTransitions = map[string][]string{
	"pending":    {"confirmed", "cancelled"},
	"confirmed":  {"processing", "cancelled"},
	"processing": {"shipped", "cancelled"},
	"shipped":    {"delivered"},
	"delivered":  {"refunded"},
	"cancelled":  {},
	"refunded":   {},
}

// OrderService handles business logic for orders.
type OrderService struct {
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
	userRepo    *repository.UserRepository
	email       *EmailService
	cfg         *config.Config
}

// NewOrderService creates a new OrderService instance.
func NewOrderService(
	orderRepo *repository.OrderRepository,
	cartRepo *repository.CartRepository,
	productRepo *repository.ProductRepository,
	userRepo *repository.UserRepository,
	email *EmailService,
	cfg *config.Config,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
		email:       email,
		cfg:         cfg,
	}
}

// Demo build: only Cash on Delivery is accepted. No payment gateway is wired
// up yet, so any other method would let an order reach "confirmed" without a
// real charge. Update this allowlist when a real PSP (Stripe, etc.) is added.
var allowedPaymentMethods = map[string]bool{
	"cod": true,
}

// CreateFromCart creates an order from the user's current cart items.
func (s *OrderService) CreateFromCart(buyerID uuid.UUID, req CreateOrderRequest) (*models.Order, error) {
	if req.PaymentMethod == "" {
		req.PaymentMethod = "cod"
	}
	if !allowedPaymentMethods[req.PaymentMethod] {
		return nil, fmt.Errorf("payment method '%s' is not supported (demo mode: cash on delivery only)", req.PaymentMethod)
	}

	cartItems, err := s.cartRepo.GetByUser(buyerID)
	if err != nil {
		return nil, fmt.Errorf("order_service: failed to get cart: %w", err)
	}
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Calculate totals and build order items.
	var subtotal float64
	orderItems := make([]models.OrderItem, 0, len(cartItems))

	for _, ci := range cartItems {
		product := ci.Product
		if product.ID == uuid.Nil {
			continue
		}

		// Verify the product is still available and has enough stock.
		if !product.IsPublished {
			return nil, fmt.Errorf("product '%s' is no longer available", product.Title)
		}
		if product.Stock < ci.Quantity {
			return nil, fmt.Errorf("insufficient stock for '%s' (available: %d, requested: %d)",
				product.Title, product.Stock, ci.Quantity)
		}

		lineTotal := product.Price * float64(ci.Quantity)
		subtotal += lineTotal

		productID := product.ID
		sellerID := product.SellerID
		orderItems = append(orderItems, models.OrderItem{
			ProductID: &productID,
			SellerID:  &sellerID,
			Title:     product.Title,
			Price:     product.Price,
			Quantity:  ci.Quantity,
			Thumbnail: product.Thumbnail,
			Status:    "pending",
		})
	}

	// Calculate tax (simple percentage) and total.
	taxRate := 0.0 // Could come from config or be region-based.
	tax := math.Round(subtotal*taxRate*100) / 100
	total := math.Round((subtotal+tax)*100) / 100

	orderNumber := s.orderRepo.GenerateOrderNumber()

	order := &models.Order{
		BuyerID:       buyerID,
		OrderNumber:   orderNumber,
		Status:        "pending",
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         total,
		PaymentStatus: "pending",
		Items:         orderItems,
	}

	// Set shipping info.
	if req.ShippingName != "" {
		order.ShippingName = &req.ShippingName
	}
	if req.ShippingAddressLine1 != "" {
		order.ShippingAddressLine1 = &req.ShippingAddressLine1
	}
	if req.ShippingAddressLine2 != "" {
		order.ShippingAddressLine2 = &req.ShippingAddressLine2
	}
	if req.ShippingCity != "" {
		order.ShippingCity = &req.ShippingCity
	}
	if req.ShippingState != "" {
		order.ShippingState = &req.ShippingState
	}
	if req.ShippingCountry != "" {
		order.ShippingCountry = &req.ShippingCountry
	}
	if req.ShippingZip != "" {
		order.ShippingZip = &req.ShippingZip
	}
	if req.ShippingPhone != "" {
		order.ShippingPhone = &req.ShippingPhone
	}
	if req.PaymentMethod != "" {
		order.PaymentMethod = &req.PaymentMethod
	}
	if req.Notes != "" {
		order.Notes = &req.Notes
	}

	// Create the order (this also decrements stock in a transaction).
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("order_service: failed to create order: %w", err)
	}

	// Clear the cart after successful order creation.
	if err := s.cartRepo.Clear(buyerID); err != nil {
		slog.Warn("order_service: failed to clear cart", "user_id", buyerID, "error", err)
	}

	// Fetch the complete order with preloaded associations.
	fullOrder, err := s.orderRepo.FindByID(order.ID)
	if err != nil {
		return nil, err
	}

	// Fire-and-forget emails. Never fail the order on email errors.
	go s.sendOrderConfirmation(buyerID, fullOrder)
	go s.notifySellersOfNewOrder(fullOrder)

	return fullOrder, nil
}

// sendOrderConfirmation sends the buyer a confirmation email for a new order.
// Runs in a goroutine; errors are logged, never returned.
func (s *OrderService) sendOrderConfirmation(buyerID uuid.UUID, order *models.Order) {
	if s.email == nil {
		return
	}
	buyer, err := s.userRepo.FindByID(buyerID)
	if err != nil || buyer == nil {
		slog.Warn("order_service: cannot resolve buyer for confirmation email", "user_id", buyerID, "error", err)
		return
	}
	if err := s.email.SendOrderConfirmation(context.Background(), buyer.Email, buyer.FullName, order); err != nil {
		slog.Warn("order_service: order confirmation email failed", "order", order.OrderNumber, "error", err)
	}
}

// notifySellersOfNewOrder sends each seller in an order their own breakdown of
// the items they need to fulfill. Runs in a goroutine; errors are logged.
func (s *OrderService) notifySellersOfNewOrder(order *models.Order) {
	if s.email == nil {
		return
	}
	bySeller := make(map[uuid.UUID][]models.OrderItem)
	for _, item := range order.Items {
		if item.SellerID == nil {
			continue
		}
		bySeller[*item.SellerID] = append(bySeller[*item.SellerID], item)
	}
	for sellerID, items := range bySeller {
		seller, err := s.userRepo.FindByID(sellerID)
		if err != nil || seller == nil {
			slog.Warn("order_service: seller lookup failed", "seller_id", sellerID, "error", err)
			continue
		}
		if err := s.email.SendSellerNewOrder(context.Background(), seller.Email, seller.FullName, order, items); err != nil {
			slog.Warn("order_service: seller new-order email failed",
				"seller_id", sellerID, "order", order.OrderNumber, "error", err)
		}
	}
}

// GetByID retrieves an order by its ID.
func (s *OrderService) GetByID(id uuid.UUID) (*models.Order, error) {
	return s.orderRepo.FindByID(id)
}

// ListByBuyer returns a paginated list of orders for a specific buyer.
func (s *OrderService) ListByBuyer(buyerID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.orderRepo.ListByBuyer(buyerID, page, perPage)
}

// ListBySeller returns a paginated list of orders containing items from a seller.
func (s *OrderService) ListBySeller(sellerID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.orderRepo.ListBySeller(sellerID, page, perPage)
}

// UpdateStatus changes the status of an order, validating the state transition.
func (s *OrderService) UpdateStatus(orderID uuid.UUID, status string, userID uuid.UUID, role string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Authorization: admins can update any order; sellers can update orders containing their items;
	// buyers can only cancel their own pending orders.
	switch role {
	case "admin":
		// Admin can update any order status.
	case "seller":
		// Verify the seller has items in this order.
		hasItems := false
		for _, item := range order.Items {
			if item.SellerID != nil && *item.SellerID == userID {
				hasItems = true
				break
			}
		}
		if !hasItems {
			return fmt.Errorf("not authorized to update this order")
		}
	case "buyer":
		if order.BuyerID != userID {
			return fmt.Errorf("not authorized to update this order")
		}
		if status != "cancelled" {
			return fmt.Errorf("buyers can only cancel orders")
		}
	default:
		return fmt.Errorf("invalid role")
	}

	// Validate the state transition.
	allowed, ok := validStatusTransitions[order.Status]
	if !ok {
		return fmt.Errorf("invalid current order status: %s", order.Status)
	}

	valid := false
	for _, allowedStatus := range allowed {
		if allowedStatus == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cannot transition from '%s' to '%s'", order.Status, status)
	}

	if err := s.orderRepo.UpdateStatus(orderID, status); err != nil {
		return err
	}

	// Fetch the updated order and notify the buyer.
	updated, err := s.orderRepo.FindByID(orderID)
	if err == nil && updated != nil {
		go s.sendOrderStatusUpdate(updated)
	}

	return nil
}

// sendOrderStatusUpdate emails the buyer when their order's status changes.
func (s *OrderService) sendOrderStatusUpdate(order *models.Order) {
	if s.email == nil {
		return
	}
	buyer, err := s.userRepo.FindByID(order.BuyerID)
	if err != nil || buyer == nil {
		return
	}
	if err := s.email.SendOrderStatusUpdate(context.Background(), buyer.Email, buyer.FullName, order); err != nil {
		slog.Warn("order_service: status-update email failed", "order", order.OrderNumber, "error", err)
	}
}

// GetAdminStats returns aggregate order statistics for the admin dashboard.
func (s *OrderService) GetAdminStats() (*repository.OrderStats, error) {
	return s.orderRepo.GetStats()
}

// GetSellerStats returns the authoritative stats panel for a seller dashboard.
func (s *OrderService) GetSellerStats(sellerID uuid.UUID) (*repository.SellerStats, error) {
	return s.orderRepo.GetSellerStats(sellerID)
}
