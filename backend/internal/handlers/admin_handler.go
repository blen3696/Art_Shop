package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artshop/backend/internal/services"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AdminHandler handles HTTP requests for admin-only endpoints.
type AdminHandler struct {
	adminService *services.AdminService
}

// NewAdminHandler creates a new AdminHandler instance.
func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// Dashboard handles GET /api/admin/dashboard (requires admin).
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get dashboard stats")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

// ListUsers handles GET /api/admin/users (requires admin).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	role := q.Get("role")

	users, total, err := h.adminService.ListUsers(page, perPage, role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list users")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, users, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// UpdateUserRole handles PUT /api/admin/users/:id/role (requires admin).
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Role == "" {
		response.ValidationError(w, map[string]string{
			"role": "Role is required",
		})
		return
	}

	if err := h.adminService.UpdateUserRole(userID, req.Role); err != nil {
		response.Error(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "User role updated successfully",
	})
}

// ToggleUserActive handles PUT /api/admin/users/:id/toggle-active (requires admin).
func (h *AdminHandler) ToggleUserActive(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	if err := h.adminService.ToggleUserActive(userID); err != nil {
		response.Error(w, http.StatusBadRequest, "TOGGLE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "User active status toggled",
	})
}

// ListOrders handles GET /api/admin/orders (requires admin).
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	status := q.Get("status")

	orders, total, err := h.adminService.ListAllOrders(page, perPage, status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list orders")
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

// ToggleProductFeatured handles PUT /api/admin/products/:id/toggle-featured (requires admin).
func (h *AdminHandler) ToggleProductFeatured(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if err := h.adminService.ToggleProductFeatured(productID); err != nil {
		response.Error(w, http.StatusBadRequest, "TOGGLE_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Product featured status toggled",
	})
}

// Revenue handles GET /api/admin/revenue (requires admin).
func (h *AdminHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.adminService.GetRevenueByMonth()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get revenue data")
		return
	}

	response.JSON(w, http.StatusOK, revenue)
}
