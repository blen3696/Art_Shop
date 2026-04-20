package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/models"
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

// UpdateProfile handles PUT /api/auth/profile (requires auth).
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// ForgotPassword handles POST /api/auth/forgot-password (public).
// Always returns 200 with a generic message — never reveals whether the
// email is registered (prevents account enumeration).
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Email == "" {
		response.ValidationError(w, map[string]string{"email": "Email is required"})
		return
	}

	// Errors are logged inside the service; we never surface them to the client.
	_ = h.authService.RequestPasswordReset(req.Email)

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, a reset link is on its way.",
	})
}

// ResetPassword handles POST /api/auth/reset-password (public).
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	errs := make(map[string]string)
	if req.Token == "" {
		errs["token"] = "Token is required"
	}
	if req.NewPassword == "" {
		errs["new_password"] = "New password is required"
	} else if len(req.NewPassword) < 8 {
		errs["new_password"] = "New password must be at least 8 characters"
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.Error(w, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully. You can now sign in.",
	})
}

// ChangePassword handles POST /api/auth/change-password (requires auth).
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	errs := make(map[string]string)
	if req.CurrentPassword == "" {
		errs["current_password"] = "Current password is required"
	}
	if req.NewPassword == "" {
		errs["new_password"] = "New password is required"
	} else if len(req.NewPassword) < 8 {
		errs["new_password"] = "New password must be at least 8 characters"
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	if err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.Error(w, http.StatusBadRequest, "PASSWORD_CHANGE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
