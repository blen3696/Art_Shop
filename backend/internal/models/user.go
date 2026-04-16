package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core users table supporting buyer, seller, and admin roles.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	FullName     string    `gorm:"not null;size:255" json:"full_name"`
	Role         string    `gorm:"size:20;default:buyer" json:"role"`
	AvatarURL    *string   `json:"avatar_url"`
	Phone        *string   `gorm:"size:20" json:"phone"`
	Bio          *string   `json:"bio"`
	AddressLine1 *string   `gorm:"size:255" json:"address_line1"`
	AddressLine2 *string   `gorm:"size:255" json:"address_line2"`
	City         *string   `gorm:"size:100" json:"city"`
	State        *string   `gorm:"size:100" json:"state"`
	Country      *string   `gorm:"size:100" json:"country"`
	ZipCode      *string   `gorm:"size:20" json:"zip_code"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relations
	SellerProfile *SellerProfile `gorm:"foreignKey:UserID" json:"seller_profile,omitempty"`
}

// TableName overrides the default table name.
func (User) TableName() string {
	return "users"
}

// SellerProfile stores extended information for users with role='seller'.
type SellerProfile struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	StoreName        string    `gorm:"not null;size:255" json:"store_name"`
	StoreDescription *string   `json:"store_description"`
	LogoURL          *string   `json:"logo_url"`
	BannerURL        *string   `json:"banner_url"`
	IsVerified       bool      `gorm:"default:false" json:"is_verified"`
	TotalSales       int       `gorm:"default:0" json:"total_sales"`
	TotalRevenue     float64   `gorm:"type:decimal(12,2);default:0" json:"total_revenue"`
	Rating           float64   `gorm:"type:decimal(3,2);default:0" json:"rating"`
	CommissionRate   float64   `gorm:"type:decimal(4,2);default:10.00" json:"commission_rate"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName overrides the default table name.
func (SellerProfile) TableName() string {
	return "seller_profiles"
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
}

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// RefreshTokenRequest is used to obtain a new access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateProfileRequest allows users to update their profile fields.
type UpdateProfileRequest struct {
	FullName     *string `json:"full_name"`
	Phone        *string `json:"phone"`
	Bio          *string `json:"bio"`
	AvatarURL    *string `json:"avatar_url"`
	AddressLine1 *string `json:"address_line1"`
	AddressLine2 *string `json:"address_line2"`
	City         *string `json:"city"`
	State        *string `json:"state"`
	Country      *string `json:"country"`
	ZipCode      *string `json:"zip_code"`
}

// CreateSellerProfileRequest is used when a buyer upgrades to seller.
type CreateSellerProfileRequest struct {
	StoreName        string  `json:"store_name" validate:"required,min=2"`
	StoreDescription *string `json:"store_description"`
}

// UserResponse is a safe public representation of a user (no password hash).
type UserResponse struct {
	ID           uuid.UUID      `json:"id"`
	Email        string         `json:"email"`
	FullName     string         `json:"full_name"`
	Role         string         `json:"role"`
	AvatarURL    *string        `json:"avatar_url"`
	Phone        *string        `json:"phone"`
	Bio          *string        `json:"bio"`
	AddressLine1 *string        `json:"address_line1"`
	AddressLine2 *string        `json:"address_line2"`
	City         *string        `json:"city"`
	State        *string        `json:"state"`
	Country      *string        `json:"country"`
	ZipCode      *string        `json:"zip_code"`
	IsVerified   bool           `json:"is_verified"`
	IsActive     bool           `json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	SellerProfile *SellerProfile `json:"seller_profile,omitempty"`
}

// ToResponse converts a User model to a safe public DTO.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		FullName:      u.FullName,
		Role:          u.Role,
		AvatarURL:     u.AvatarURL,
		Phone:         u.Phone,
		Bio:           u.Bio,
		AddressLine1:  u.AddressLine1,
		AddressLine2:  u.AddressLine2,
		City:          u.City,
		State:         u.State,
		Country:       u.Country,
		ZipCode:       u.ZipCode,
		IsVerified:    u.IsVerified,
		IsActive:      u.IsActive,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		SellerProfile: u.SellerProfile,
	}
}
