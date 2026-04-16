package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims extends jwt.RegisteredClaims with application-specific fields.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateTokenPair creates a signed access token and refresh token for the
// given user. The access token is short-lived (used for API calls) while the
// refresh token is long-lived (used to obtain new access tokens).
func GenerateTokenPair(
	userID uuid.UUID,
	role string,
	secret string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) (accessToken string, refreshToken string, err error) {
	accessToken, err = generateToken(userID, role, "access", secret, accessExpiry)
	if err != nil {
		return "", "", fmt.Errorf("jwt: failed to generate access token: %w", err)
	}

	refreshToken, err = generateToken(userID, role, "refresh", secret, refreshExpiry)
	if err != nil {
		return "", "", fmt.Errorf("jwt: failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a JWT string. It returns the embedded
// claims on success or an error if the token is malformed, expired, or has an
// invalid signature.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: token parse error: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token claims")
	}

	return claims, nil
}

// generateToken creates a single signed JWT with the given parameters.
func generateToken(
	userID uuid.UUID,
	role string,
	tokenType string,
	secret string,
	expiry time.Duration,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "artshop",
			Subject:   userID.String(),
			ID:        uuid.New().String(), // unique token ID (jti) for revocation support
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}
