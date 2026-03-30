package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields.
	errors := make(map[string]string)
	if req.Email == "" {
		errors["email"] = "Email is required"
	}
	if req.Password == "" {
		errors["password"] = "Password is required"
	} else if len(req.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}
	if req.FullName == "" {
		errors["full_name"] = "Full name is required"
	}
	if len(errors) > 0 {
		response.ValidationError(w, errors)
		return
	}

	authResp, err := h.authService.Register(req)
	if err != nil {
		response.Error(w, http.StatusConflict, "REGISTRATION_FAILED", err.Error())
		return
	}

	response.Created(w, authResp)
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req services.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	errors := make(map[string]string)
	if req.Email == "" {
		errors["email"] = "Email is required"
	}
	if req.Password == "" {
		errors["password"] = "Password is required"
	}
	if len(errors) > 0 {
		response.ValidationError(w, errors)
		return
	}

	authResp, err := h.authService.Login(req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, authResp)
}

// RefreshToken handles POST /api/auth/refresh.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Refresh token is required")
		return
	}

	authResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "REFRESH_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, authResp)
}

// RegisterSeller handles POST /api/auth/register-seller (requires auth).
func (h *AuthHandler) RegisterSeller(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req services.SellerRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.StoreName == "" {
		response.ValidationError(w, map[string]string{
			"store_name": "Store name is required",
		})
		return
	}

	if err := h.authService.RegisterAsSeller(userID, req); err != nil {
		response.Error(w, http.StatusBadRequest, "SELLER_REGISTRATION_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Successfully registered as seller",
	})
}

// Me handles GET /api/auth/me (requires auth).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
