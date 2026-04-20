package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/models"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// passwordResetTokenTTL is how long a reset link stays valid.
const passwordResetTokenTTL = 60 * time.Minute

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
	userRepo  *repository.UserRepository
	resetRepo *repository.PasswordResetTokenRepository
	email     *EmailService
	cfg       *config.Config
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(
	userRepo *repository.UserRepository,
	resetRepo *repository.PasswordResetTokenRepository,
	email *EmailService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		resetRepo: resetRepo,
		email:     email,
		cfg:       cfg,
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

	// Fire-and-forget welcome email.
	if s.email != nil {
		go func() {
			if err := s.email.SendWelcome(context.Background(), user.Email, user.FullName); err != nil {
				slog.Warn("auth_service: welcome email failed", "email", user.Email, "error", err)
			}
		}()
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

// RequestPasswordReset issues a single-use reset token for the given email
// (if it exists) and emails the user a reset link. The function ALWAYS returns
// nil error regardless of whether the email exists — this prevents account
// enumeration via the forgot-password endpoint.
func (s *AuthService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil || !user.IsActive {
		// Silent: don't leak whether the address is registered.
		return nil
	}

	// Invalidate any older active tokens so only the latest link works.
	_ = s.resetRepo.InvalidateForUser(user.ID)

	rawToken, hash, err := newResetToken()
	if err != nil {
		return fmt.Errorf("auth: generate reset token: %w", err)
	}

	record := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(passwordResetTokenTTL),
	}
	if err := s.resetRepo.Create(record); err != nil {
		return fmt.Errorf("auth: store reset token: %w", err)
	}

	if s.email != nil {
		go func() {
			if err := s.email.SendPasswordReset(context.Background(), user.Email, user.FullName, rawToken); err != nil {
				slog.Warn("auth_service: password reset email failed", "email", user.Email, "error", err)
			}
		}()
	}

	return nil
}

// ResetPassword consumes a reset token and sets the user's password.
func (s *AuthService) ResetPassword(rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if rawToken == "" {
		return fmt.Errorf("invalid or expired token")
	}

	hash := hashResetToken(rawToken)
	tok, err := s.resetRepo.FindActiveByHash(hash)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	user, err := s.userRepo.FindByID(tok.UserID)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}
	user.PasswordHash = string(pwdHash)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("auth: update password: %w", err)
	}

	if err := s.resetRepo.MarkUsed(tok.ID); err != nil {
		// Non-fatal: log only. The password is already changed.
		slog.Warn("auth_service: failed to mark reset token used", "token_id", tok.ID, "error", err)
	}

	return nil
}

// newResetToken returns a (rawToken, sha256Hash) pair. Only the hash is stored.
func newResetToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(buf)
	return raw, hashResetToken(raw), nil
}

func hashResetToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
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
