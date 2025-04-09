package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration.
type Config struct {
	DBHost     string        `env:"DB_HOST,required"`
	DBPort     int           `env:"DB_PORT,required"`
	DBUser     string        `env:"DB_USER,required"`
	DBPassword string        `env:"DB_PASSWORD,required"`
	DBName     string        `env:"DB_NAME,required"`
	JWTSecret  string        `env:"JWT_SECRET,required"`       // Secret key for signing JWT tokens
	TokenTTL   time.Duration `env:"TOKEN_TTL" envDefault:"1h"` // Token time-to-live
	HTTPPort   string        `env:"HTTP_PORT" envDefault:"8080"`
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Printf("Error parsing config from environment variables: %v", err)
		return nil, err
	}
	// Ensure JWTSecret is set, as it's crucial for security
	if cfg.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set.")
	}
	return cfg, nil
}
