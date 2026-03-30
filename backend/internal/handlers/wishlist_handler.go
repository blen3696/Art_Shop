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

// WishlistHandler handles HTTP requests for wishlist endpoints.
type WishlistHandler struct {
	cartRepo *repository.CartRepository
}

// NewWishlistHandler creates a new WishlistHandler instance.
func NewWishlistHandler(cartRepo *repository.CartRepository) *WishlistHandler {
	return &WishlistHandler{cartRepo: cartRepo}
}

// GetWishlist handles GET /api/wishlist (requires auth).
func (h *WishlistHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	items, err := h.cartRepo.GetWishlist(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get wishlist")
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// AddToWishlist handles POST /api/wishlist (requires auth).
func (h *WishlistHandler) AddToWishlist(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req struct {
		ProductID string `json:"product_id"`
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

	item := &models.Wishlist{
		UserID:    userID,
		ProductID: productID,
	}

	if err := h.cartRepo.AddToWishlist(item); err != nil {
		response.Error(w, http.StatusInternalServerError, "ADD_FAILED", "Failed to add to wishlist")
		return
	}

	response.Created(w, map[string]string{
		"message": "Added to wishlist",
	})
}

// RemoveFromWishlist handles DELETE /api/wishlist/:productId (requires auth).
func (h *WishlistHandler) RemoveFromWishlist(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	productIDStr := chi.URLParam(r, "productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if err := h.cartRepo.RemoveFromWishlist(userID, productID); err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Wishlist item not found")
		return
	}

	response.NoContent(w)
}
