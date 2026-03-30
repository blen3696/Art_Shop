package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ReviewHandler handles HTTP requests for review endpoints.
type ReviewHandler struct {
	reviewRepo *repository.ReviewRepository
}

// NewReviewHandler creates a new ReviewHandler instance.
func NewReviewHandler(reviewRepo *repository.ReviewRepository) *ReviewHandler {
	return &ReviewHandler{reviewRepo: reviewRepo}
}

// GetByProduct handles GET /api/products/:id/reviews.
func (h *ReviewHandler) GetByProduct(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	reviews, total, err := h.reviewRepo.GetByProduct(productID, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get reviews")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, reviews, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// Create handles POST /api/products/:id/reviews (requires auth).
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate rating.
	if req.Rating < 1 || req.Rating > 5 {
		response.ValidationError(w, map[string]string{
			"rating": "Rating must be between 1 and 5",
		})
		return
	}

	// Check if the user has purchased this product.
	isVerified := h.reviewRepo.HasPurchased(userID, productID)

	review := &models.Review{
		ProductID:          productID,
		UserID:             userID,
		Rating:             req.Rating,
		IsVerifiedPurchase: isVerified,
	}

	if req.Title != "" {
		review.Title = &req.Title
	}
	if req.Comment != "" {
		review.Comment = &req.Comment
	}

	if err := h.reviewRepo.Create(review); err != nil {
		response.Error(w, http.StatusBadRequest, "REVIEW_FAILED", "Failed to create review. You may have already reviewed this product.")
		return
	}

	response.Created(w, review)
}

// Delete handles DELETE /api/reviews/:id (requires auth, owner only).
func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	reviewIDStr := chi.URLParam(r, "id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid review ID")
		return
	}

	if err := h.reviewRepo.Delete(reviewID, userID); err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Review not found or not owned by you")
		return
	}

	response.NoContent(w)
}
