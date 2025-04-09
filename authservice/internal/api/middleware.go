package api

import (
	"context"
	"net/http"
	"strings"

	"authservice/internal/auth"
)

// contextKey is a custom type to avoid context key collisions.
type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware creates a middleware handler for JWT authentication.
func AuthMiddleware(authService auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format (must be Bearer token)")
				return
			}

			tokenString := headerParts[1]
			claims, err := authService.VerifyToken(tokenString)
			if err != nil {
				// Log the specific error for debugging? Maybe not in production.
				// log.Printf("Token verification error: %v", err)
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			r = r.WithContext(ctx)

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}
