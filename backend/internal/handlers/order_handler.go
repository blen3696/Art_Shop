package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// OrderHandler handles HTTP requests for order endpoints.
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new OrderHandler instance.
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Create handles POST /api/orders (requires auth) -- creates an order from the cart.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req services.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required shipping fields.
	errors := make(map[string]string)
	if req.ShippingName == "" {
		errors["shipping_name"] = "Shipping name is required"
	}
	if req.ShippingAddressLine1 == "" {
		errors["shipping_address_line1"] = "Shipping address is required"
	}
	if req.ShippingCity == "" {
		errors["shipping_city"] = "City is required"
	}
	if req.ShippingCountry == "" {
		errors["shipping_country"] = "Country is required"
	}
	if len(errors) > 0 {
		response.ValidationError(w, errors)
		return
	}

	order, err := h.orderService.CreateFromCart(userID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ORDER_FAILED", err.Error())
		return
	}

	response.Created(w, order)
}

// List handles GET /api/orders (requires auth) -- lists the current user's orders.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	orders, total, err := h.orderService.ListByBuyer(userID, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, orders, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/orders/:id (requires auth).
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	role := middleware.GetUserRoleFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid order ID")
		return
	}

	order, err := h.orderService.GetByID(orderID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Order not found")
		return
	}

	// Authorization: buyers see their own orders; sellers see orders with their items; admins see all.
	if role != "admin" {
		if role == "buyer" && order.BuyerID != userID {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "Not authorized to view this order")
			return
		}
		if role == "seller" {
			hasItems := false
			for _, item := range order.Items {
				if item.SellerID != nil && *item.SellerID == userID {
					hasItems = true
					break
				}
			}
			if !hasItems && order.BuyerID != userID {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Not authorized to view this order")
				return
			}
		}
	}

	response.JSON(w, http.StatusOK, order)
}

// UpdateStatus handles PUT /api/orders/:id/status (requires seller/admin).
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	role := middleware.GetUserRoleFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Status == "" {
		response.ValidationError(w, map[string]string{
			"status": "Status is required",
		})
		return
	}

	if err := h.orderService.UpdateStatus(orderID, req.Status, userID, role); err != nil {
		response.Error(w, http.StatusBadRequest, "STATUS_UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Order status updated successfully",
	})
}

// SellerOrders handles GET /api/seller/orders (requires seller).
func (h *OrderHandler) SellerOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	orders, total, err := h.orderService.ListBySeller(userID, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, orders, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}
