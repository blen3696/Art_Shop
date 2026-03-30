package repository

import (
	"fmt"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles all database operations for users and seller profiles.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user record into the database.
func (r *UserRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("user_repo: failed to create user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their UUID, preloading the SellerProfile if present.
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("SellerProfile").First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("user_repo: user not found: %w", err)
	}
	return &user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("user_repo: user not found: %w", err)
	}
	return &user, nil
}

// Update saves changes to an existing user record.
func (r *UserRepository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("user_repo: failed to update user: %w", err)
	}
	return nil
}

// Delete removes a user by their UUID (soft or hard delete depending on GORM config).
func (r *UserRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("user_repo: failed to delete user: %w", err)
	}
	return nil
}

// List returns a paginated list of users, optionally filtered by role.
func (r *UserRepository) List(page, perPage int, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("user_repo: failed to count users: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("user_repo: failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateLastLogin sets the last_login_at timestamp to the current time.
func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&models.User{}).Where("id = ?", id).Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("user_repo: failed to update last login: %w", err)
	}
	return nil
}

// CreateSellerProfile inserts a new seller profile record.
func (r *UserRepository) CreateSellerProfile(profile *models.SellerProfile) error {
	if err := r.db.Create(profile).Error; err != nil {
		return fmt.Errorf("user_repo: failed to create seller profile: %w", err)
	}
	return nil
}

// GetSellerProfile retrieves the seller profile for a given user ID.
func (r *UserRepository) GetSellerProfile(userID uuid.UUID) (*models.SellerProfile, error) {
	var profile models.SellerProfile
	if err := r.db.First(&profile, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user_repo: seller profile not found: %w", err)
	}
	return &profile, nil
}
