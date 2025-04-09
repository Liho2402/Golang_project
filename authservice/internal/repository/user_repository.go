package repository

import (
	"context"
	"errors"
	"time"

	"authservice/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

// postgresUserRepository implements UserRepository for PostgreSQL.
type postgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL user repository.
func NewPostgresUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{pool: pool}
}

// CreateUser inserts a new user into the database.
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	query := `INSERT INTO users (email, password_hash, created_at, updated_at)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id`
	now := time.Now()
	var userID int64
	err := r.pool.QueryRow(ctx, query, user.Email, user.Password, now, now).Scan(&userID)
	if err != nil {
		// Basic check for unique constraint violation (adjust if needed based on specific error code/message)
		if errors.Is(err, pgx.ErrNoRows) { // Check pgx specific errors if necessary
			// Handle specific errors, maybe check constraint violation for email
		}
		// A more robust check might involve checking the PostgreSQL error code (e.g., 23505 for unique_violation)
		// For simplicity, we'll assume other errors are potential conflicts for now
		// This check is basic and might need refinement based on actual DB errors
		if err.Error() == "unique constraint violation" { // Very basic check
			return 0, ErrEmailExists
		}
		return 0, err
	}
	return userID, nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *postgresUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
