package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"authservice/internal/config"
	"authservice/internal/domain"
	"authservice/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents the JWT claims.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthService provides authentication related functionalities.
type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	VerifyToken(tokenString string) (*Claims, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cfg: cfg}
}

// Register creates a new user.
func (s *authService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		// Handle potential duplicate email error from repository
		if errors.Is(err, repository.ErrEmailExists) {
			return nil, err // Return the specific error
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = userID
	user.Password = "" // Clear password hash before returning
	return user, nil
}

// Login authenticates a user and returns a JWT token.
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", fmt.Errorf("invalid credentials") // Generic error for security
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// If passwords don't match
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", fmt.Errorf("invalid credentials")
		}
		// Other errors during comparison
		return "", fmt.Errorf("password comparison failed: %w", err)
	}

	// Generate JWT token
	expirationTime := time.Now().Add(s.cfg.TokenTTL)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "authservice", // Optional: identify the issuer
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken validates the JWT token and returns the claims.
func (s *authService) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
