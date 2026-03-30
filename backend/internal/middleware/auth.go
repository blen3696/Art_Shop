package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/artshop/backend/pkg/response"
	"github.com/artshop/backend/pkg/utils"
	"github.com/google/uuid"
)

// Context keys used to store authenticated user information.
type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyRole   contextKey = "user_role"
)

// RequireAuth is a chi middleware that validates the JWT bearer token in the
// Authorization header. On success it injects the user ID and role into the
// request context. On failure it returns 401 Unauthorized.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := extractBearerToken(r)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or malformed authorization header")
				return
			}

			claims, err := utils.ValidateToken(token, jwtSecret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
				return
			}

			// Only accept access tokens (not refresh tokens).
			if claims.TokenType != "access" {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN_TYPE", "Expected an access token")
				return
			}

			// Inject user information into context.
			ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a chi middleware that restricts access to users whose
// role matches one of the provided values. It must be placed after RequireAuth
// in the middleware chain.
//
// Usage:
//
//	r.With(middleware.RequireAuth(secret), middleware.RequireRole("admin")).Get("/admin", handler)
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRoleFromContext(r.Context())
			if _, ok := allowed[role]; !ok {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

// GetUserIDFromContext retrieves the authenticated user's UUID from the
// request context. Returns uuid.Nil if not present.
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

// GetUserRoleFromContext retrieves the authenticated user's role from the
// request context. Returns an empty string if not present.
func GetUserRoleFromContext(ctx context.Context) string {
	role, ok := ctx.Value(contextKeyRole).(string)
	if !ok {
		return ""
	}
	return role
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// extractBearerToken pulls the token string from the Authorization header.
// Expected format: "Bearer <token>".
func extractBearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
