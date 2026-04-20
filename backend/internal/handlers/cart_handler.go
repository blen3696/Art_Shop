package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CartHandler handles HTTP requests for cart endpoints.
type CartHandler struct {
	cartRepo *repository.CartRepository
}

// NewCartHandler creates a new CartHandler instance.
func NewCartHandler(cartRepo *repository.CartRepository) *CartHandler {
	return &CartHandler{cartRepo: cartRepo}
}

// GetCart handles GET /api/cart (requires auth). Returns the cart items array
// directly so the frontend can compute totals locally — matches the typed
// contract `CartItem[]` declared in the API client.
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	items, err := h.cartRepo.GetByUser(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get cart")
		return
	}
	if items == nil {
		items = []models.CartItem{}
	}

	response.JSON(w, http.StatusOK, items)
}

// AddItem handles POST /api/cart (requires auth).
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if req.Quantity < 1 {
		req.Quantity = 1
	}

	item := &models.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	if err := h.cartRepo.AddItem(item); err != nil {
		response.Error(w, http.StatusInternalServerError, "ADD_FAILED", "Failed to add item to cart")
		return
	}

	response.Created(w, map[string]string{
		"message": "Item added to cart",
	})
}

// UpdateQuantity handles PUT /api/cart/:productId (requires auth).
func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	productIDStr := chi.URLParam(r, "productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.cartRepo.UpdateQuantity(userID, productID, req.Quantity); err != nil {
		response.Error(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Cart item updated",
	})
}

// RemoveItem handles DELETE /api/cart/:productId (requires auth).
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	productIDStr := chi.URLParam(r, "productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if err := h.cartRepo.RemoveItem(userID, productID); err != nil {
		response.Error(w, http.StatusBadRequest, "REMOVE_FAILED", err.Error())
		return
	}

	response.NoContent(w)
}

// Clear handles DELETE /api/cart (requires auth) -- clears the entire cart.
func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.cartRepo.Clear(userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "CLEAR_FAILED", "Failed to clear cart")
		return
	}

	response.NoContent(w)
}
