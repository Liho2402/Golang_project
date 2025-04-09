package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"authservice/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates a new PostgreSQL connection pool.
func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Printf("Unable to parse DSN: %v\n", err)
		return nil, err
	}

	// Set connection pool parameters (adjust as needed)
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	log.Println("Connecting to database...")
	var pool *pgxpool.Pool
	// err is already declared above by poolConfig, err := ...
	// Retry connection logic (useful for docker-compose startup order)
	for i := 0; i < 5; i++ {
		// Use := to declare pool in this scope, but only assign to err
		var tempPool *pgxpool.Pool
		tempPool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			// Check the connection
			pingErr := tempPool.Ping(context.Background())
			if pingErr == nil {
				log.Println("Database connection established successfully.")
				pool = tempPool // Assign to the outer pool variable on success
				return pool, nil
			}
			log.Printf("Database ping failed: %v. Retrying... (%d/5)", pingErr, i+1)
			tempPool.Close() // Close the temporary pool if ping fails
		} else {
			log.Printf("Unable to create connection pool: %v. Retrying... (%d/5)", err, i+1)
		}
		time.Sleep(2 * time.Second)
	}

	log.Printf("Failed to connect to database after multiple retries.")
	// Return the last error encountered (could be from NewWithConfig or Ping)
	return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
}
