package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// respondAIError maps AI service errors to the right HTTP status: 503 when
// AI is disabled (no key configured) so the frontend can hide features
// gracefully, 500 for actual provider failures.
func respondAIError(w http.ResponseWriter, err error, action string) {
	if errors.Is(err, services.ErrAIDisabled) {
		response.Error(w, http.StatusServiceUnavailable, "AI_DISABLED",
			"AI features are not configured on this server.")
		return
	}
	response.Error(w, http.StatusInternalServerError, "AI_ERROR", "Failed to "+action+": "+err.Error())
}

// AIHandler handles HTTP requests for AI-powered endpoints.
type AIHandler struct {
	aiService *services.AIService
	db        *gorm.DB
}

// NewAIHandler creates a new AIHandler instance.
func NewAIHandler(aiService *services.AIService, db *gorm.DB) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		db:        db,
	}
}

// GenerateDescription handles POST /api/ai/generate-description (requires seller).
func (h *AIHandler) GenerateDescription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title      string   `json:"title"`
		Medium     string   `json:"medium"`
		Dimensions string   `json:"dimensions"`
		Tags       []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Title == "" {
		response.ValidationError(w, map[string]string{
			"title": "Title is required",
		})
		return
	}

	description, err := h.aiService.GenerateProductDescription(req.Title, req.Medium, req.Dimensions, req.Tags)
	if err != nil {
		respondAIError(w, err, "generate description")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"description": description,
	})
}

// GenerateTags handles POST /api/ai/generate-tags (requires seller).
func (h *AIHandler) GenerateTags(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Title == "" {
		response.ValidationError(w, map[string]string{
			"title": "Title is required",
		})
		return
	}

	tags, err := h.aiService.GenerateProductTags(req.Title, req.Description)
	if err != nil {
		respondAIError(w, err, "generate tags")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"tags": tags,
	})
}

// Recommendations handles GET /api/ai/recommendations (requires auth).
func (h *AIHandler) Recommendations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	// Fetch the user's recent browsing history.
	type historyRow struct {
		ProductID uuid.UUID
	}
	var history []historyRow
	h.db.Table("browsing_history").
		Select("product_id").
		Where("user_id = ?", userID).
		Order("viewed_at DESC").
		Limit(50).
		Scan(&history)

	browsingHistory := make([]uuid.UUID, len(history))
	for i, h := range history {
		browsingHistory[i] = h.ProductID
	}

	recommendations, err := h.aiService.GetRecommendations(userID, browsingHistory, h.db)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "AI_ERROR", "Failed to get recommendations")
		return
	}

	response.JSON(w, http.StatusOK, recommendations)
}
