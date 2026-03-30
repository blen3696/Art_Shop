package handlers

import (
	"net/http"
	"strconv"

	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationHandler handles HTTP requests for notification endpoints.
type NotificationHandler struct {
	db *gorm.DB
}

// NewNotificationHandler creates a new NotificationHandler instance.
func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// List handles GET /api/notifications (requires auth).
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	var notifications []models.Notification
	var total int64

	query := h.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to count notifications")
		return
	}

	offset := (page - 1) * perPage
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&notifications).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get notifications")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	response.Paginated(w, http.StatusOK, notifications, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// MarkAsRead handles PUT /api/notifications/:id/read (requires auth).
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	notifID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		return
	}

	result := h.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true)

	if result.Error != nil {
		response.Error(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to mark notification as read")
		return
	}
	if result.RowsAffected == 0 {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PUT /api/notifications/read-all (requires auth).
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to mark all notifications as read")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "All notifications marked as read",
	})
}

// UnreadCount handles GET /api/notifications/unread-count (requires auth).
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	var count int64
	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get unread count")
		return
	}

	response.JSON(w, http.StatusOK, map[string]int64{
		"unread_count": count,
	})
}
