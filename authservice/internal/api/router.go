package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates a new chi router and sets up routes.
func NewRouter(authHandler *AuthHandler) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)    // Log requests
	r.Use(middleware.Recoverer) // Recover from panics
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.StripSlashes) // Strip trailing slashes

	// Public routes
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	// Example of a protected route (requires VerifyToken implementation in middleware)
	// r.Group(func(r chi.Router) {
	// 	r.Use(AuthMiddleware(authService)) // Assuming you create an AuthMiddleware
	// 	r.Get("/profile", GetUserProfileHandler)
	// })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Auth service is healthy!"))
	})

	// Protected routes (require valid JWT)
	r.Group(func(r chi.Router) {
		// Apply the AuthMiddleware using the authService from the handler
		r.Use(AuthMiddleware(authHandler.authService))

		// Define protected endpoints here
		r.Get("/me", authHandler.GetUserProfile)
		// Add other protected routes like /change-password, /update-profile etc.
	})

	return r
}
