package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"authservice/internal/auth"
	"authservice/internal/repository" // For error checking
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authService auth.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest defines the expected JSON body for registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest defines the expected JSON body for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines the JSON response for successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse defines the standard JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles user registration requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Basic validation (consider adding more robust validation)
	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters long")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			respondWithError(w, http.StatusConflict, "Email already exists")
		} else {
			log.Printf("Error registering user: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, user) // Return the created user (excluding password)
}

// Login handles user login requests.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" { // Check for specific error string from service
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			log.Printf("Error logging in user: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to login")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// GetUserProfile is a protected handler that retrieves the user ID from the context.
func (h *AuthHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		// This should ideally not happen if middleware is correctly applied
		log.Printf("Error: User ID not found in context or not an int64")
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user information")
		return
	}

	// In a real application, you would fetch user details from the repository using userID
	// user, err := h.userRepo.GetUserByID(r.Context(), userID) // Assuming userRepo is added to AuthHandler
	// if err != nil { ... handle error ... }

	// For now, just return the ID
	response := map[string]interface{}{"user_id": userID}
	respondWithJSON(w, http.StatusOK, response)
}

// respondWithError sends a JSON error response.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON sends a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
