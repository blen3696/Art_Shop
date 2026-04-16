package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest holds the data required to register a new user.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// LoginRequest holds the data required to log in.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SellerRegisterRequest holds the data required to register as a seller.
type SellerRegisterRequest struct {
	StoreName        string `json:"store_name"`
	StoreDescription string `json:"store_description"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// AuthService handles authentication and authorization logic.
type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register creates a new user account, hashes the password, and generates tokens.
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if email is already taken.
	_, err := s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to hash password: %w", err)
	}

	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Phone:        phone,
		Role:         "buyer",
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("auth: failed to create user: %w", err)
	}

	// Generate token pair.
	accessExpiry := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	refreshExpiry := time.Duration(s.cfg.JWTRefreshExpiryHours) * time.Hour

	accessToken, refreshToken, err := utils.GenerateTokenPair(
		user.ID, user.Role, s.cfg.JWTSecret, accessExpiry, refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login verifies credentials, updates the last login timestamp, and returns tokens.
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login timestamp (best effort; do not block on error).
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// Generate token pair.
	accessExpiry := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	refreshExpiry := time.Duration(s.cfg.JWTRefreshExpiryHours) * time.Hour

	accessToken, refreshToken, err := utils.GenerateTokenPair(
		user.ID, user.Role, s.cfg.JWTSecret, accessExpiry, refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate tokens: %w", err)
	}

	// Re-fetch to get the seller profile if present.
	user, _ = s.userRepo.FindByID(user.ID)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Generate a new token pair.
	accessExpiry := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	refreshExpiry := time.Duration(s.cfg.JWTRefreshExpiryHours) * time.Hour

	newAccess, newRefresh, err := utils.GenerateTokenPair(
		user.ID, user.Role, s.cfg.JWTSecret, accessExpiry, refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}

// RegisterAsSeller upgrades a buyer account to a seller, creating a seller profile.
func (s *AuthService) RegisterAsSeller(userID uuid.UUID, req SellerRegisterRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.Role == "seller" || user.Role == "admin" {
		// Check if profile already exists.
		_, err := s.userRepo.GetSellerProfile(userID)
		if err == nil {
			return fmt.Errorf("seller profile already exists")
		}
		if user.Role == "admin" {
			// Admin can also have a seller profile; don't change role.
		}
	}

	// Update user role to seller (unless admin).
	if user.Role != "admin" {
		user.Role = "seller"
		if err := s.userRepo.Update(user); err != nil {
			return fmt.Errorf("auth: failed to update user role: %w", err)
		}
	}

	var storeDesc *string
	if req.StoreDescription != "" {
		storeDesc = &req.StoreDescription
	}

	profile := &models.SellerProfile{
		UserID:           userID,
		StoreName:        req.StoreName,
		StoreDescription: storeDesc,
	}

	if err := s.userRepo.CreateSellerProfile(profile); err != nil {
		// Check if it's a duplicate key error (profile already exists).
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("seller profile already exists")
		}
		return fmt.Errorf("auth: failed to create seller profile: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID (used by the /me endpoint).
func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// UpdateProfile updates the authenticated user's profile fields.
func (s *AuthService) UpdateProfile(userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.AddressLine1 != nil {
		user.AddressLine1 = req.AddressLine1
	}
	if req.AddressLine2 != nil {
		user.AddressLine2 = req.AddressLine2
	}
	if req.City != nil {
		user.City = req.City
	}
	if req.State != nil {
		user.State = req.State
	}
	if req.Country != nil {
		user.Country = req.Country
	}
	if req.ZipCode != nil {
		user.ZipCode = req.ZipCode
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.userRepo.FindByID(userID)
}

// ChangePassword validates the current password and sets a new one.
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
