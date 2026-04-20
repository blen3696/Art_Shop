package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken stores the SHA-256 hash of a password reset token.
// The raw token is only ever in the email link — never in the database.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
