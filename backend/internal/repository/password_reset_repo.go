package repository

import (
	"fmt"
	"time"

	"github.com/artshop/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordResetTokenRepository handles persistence for password reset tokens.
type PasswordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// Create stores a new (already-hashed) reset token.
func (r *PasswordResetTokenRepository) Create(t *models.PasswordResetToken) error {
	if err := r.db.Create(t).Error; err != nil {
		return fmt.Errorf("password_reset_repo: create: %w", err)
	}
	return nil
}

// FindActiveByHash returns a token if it exists, has not been used, and has
// not expired. Anything else returns gorm.ErrRecordNotFound.
func (r *PasswordResetTokenRepository) FindActiveByHash(hash string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := r.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now()).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MarkUsed sets used_at on a token so it can't be replayed.
func (r *PasswordResetTokenRepository) MarkUsed(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&models.PasswordResetToken{}).
		Where("id = ?", id).
		Update("used_at", now).Error; err != nil {
		return fmt.Errorf("password_reset_repo: mark used: %w", err)
	}
	return nil
}

// InvalidateForUser marks every active token for a user as used. Called when
// issuing a new token so previous links stop working.
func (r *PasswordResetTokenRepository) InvalidateForUser(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}
