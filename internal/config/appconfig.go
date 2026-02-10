package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Server
	Port            string `env:"PORT" envDefault:"8080"`
	AllowedOrigings string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`

	// Database Configuration
	PostgresHost         string `env:"POSTGRES_HOST"`
	PostgresPort         string `env:"POSTGRES_PORT"`
	PostgresDB           string `env:"POSTGRES_DB"`
	PostgresUser         string `env:"POSTGRES_USER"`
	PostgresPasswordFile string `env:"POSTGRES_PASSWORD_FILE"`
	PostgresSSLMode      string `env:"POSTGRES_SSL_MODE"`
	SecretsPath          string `env:"SECRETS_PATH"`

	// Admin auth
	AdminUsername     string `env:"ADMIN_USERNAME"`
	AdminPasswordFile string `env:"ADMIN_PASSWORD_FILE"`
	JWTSecretFile     string `env:"JWT_SECRET_FILE"`

	FrontendURL string `env:"FRONTEND_URL"`
	LogLevel    string `env:"LOG_LEVEL"`
}

func Load() (*Config, error) {
	// Load the env file
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Failed to find or load .env file")
	}

	cfg := &Config{}

	// Parse the env variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate required configurations
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate configs: %w", err)
	}

	log.Info().
		Str("port:", cfg.PostgresPort).
		Str("database:", cfg.PostgresDB).
		Str("log level:", cfg.LogLevel).
		Msg("Configuration loaded succesfully")

	return cfg, nil
}

// Helper func to validate configurations
func (c *Config) Validate() error {
	if c.PostgresHost == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}
	if c.PostgresPort == "" {
		return fmt.Errorf("POSTGRES_PORT is required")
	}
	if c.PostgresDB == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}
	if c.PostgresUser == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.PostgresPasswordFile == "" {
		return fmt.Errorf("POSTGRES_PASSWORD_FILE is required")
	}
	if !filepath.IsAbs(c.PostgresPasswordFile) && c.SecretsPath == "" {
		return fmt.Errorf("SECRETS_PATH is required when using relative paths for POSTGRES_PASSWORD_FILE")
	}
	if c.AdminUsername == "" {
		return fmt.Errorf("ADMIN_USERNAME is required")
	}
	if c.AdminPasswordFile == "" {
		return fmt.Errorf("ADMIN_PASSWORD_FILE is required")
	}
	if c.JWTSecretFile == "" {
		return fmt.Errorf("JWT_SECRET_FILE is required")
	}

	return nil
}

// Constructs the database URL from individual components
func (c *Config) GetDatabaseURL() (string, error) {
	password, err := c.GetDatbasePassword()
	if err != nil {
		return "", fmt.Errorf("failed to get the datbase password: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser, password, c.PostgresHost, c.PostgresPort, c.PostgresDB, c.PostgresSSLMode)

	log.Info().
		Str("host", c.PostgresHost).
		Str("port", c.PostgresPort).
		Str("database", c.PostgresDB).
		Str("user", c.PostgresUser).
		Str("ssl_mode", c.PostgresSSLMode).
		Msg("Constructed database URL from individual components")

	return dbURL, nil
}

// Helper func to get db password from secrets file
func (c *Config) GetDatbasePassword() (string, error) {
	if c.PostgresPasswordFile == "" {
		return "", fmt.Errorf("POSTGRES_PASSWORD_FILE is required")
	}

	filePath := c.PostgresPasswordFile

	if !filepath.IsAbs(filePath) && c.SecretsPath != "" {
		filePath = filepath.Join(c.SecretsPath, filePath)
		log.Debug().
			Str("relative_path", c.PostgresPasswordFile).
			Str("secrets_path", c.SecretsPath).
			Str("full_path", filePath).
			Msg("Using relative path with SECRETS_PATH")
	} else if !filepath.IsAbs(filePath) && c.SecretsPath == "" {
		return "", fmt.Errorf("relative path provided for POSTGRES_PASSWORD_FILE but SECRETS_PATH is not set")
	}

	passwordBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read the database password file from %s:%w", filePath, err)
	}
	password := string(passwordBytes)
	password = strings.TrimSpace(password)

	log.Info().Str("password_file", filePath).Msg("using POSTGRES_PASSWORD from file")

	return password, nil
}

// Helper func to get JWT from secrets file
func (c *Config) GetJWTSecret() (string, error) {
	filePath := c.JWTSecretFile

	if !filepath.IsAbs(filePath) && c.SecretsPath != "" {
		filePath = filepath.Join(c.SecretsPath, filePath)
		log.Debug().
			Str("relative_path", c.PostgresPasswordFile).
			Str("secrets_path", c.SecretsPath).
			Str("full_path", filePath).
			Msg("Using relative path with SECRETS_PATH")
	} else if !filepath.IsAbs(filePath) && c.SecretsPath == "" {
		return "", fmt.Errorf("relative path provided for JWT_SECRET_FILE but SECRETS_PATH is not set")
	}

	secretBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read JWT secret file from %s: %w", filePath, err)
	}

	return strings.TrimSpace(string(secretBytes)), nil
}

// Reads the admin password hash
func (c *Config) GetAdminPasswordHash() (string, error) {
	if c.AdminPasswordFile == "" {
		return "", fmt.Errorf("ADMIN_PASSWORD_FILE is required")
	}

	filePath := c.AdminPasswordFile

	if !filepath.IsAbs(filePath) && c.SecretsPath != "" {
		filePath = filepath.Join(c.SecretsPath, filePath)
	} else if !filepath.IsAbs(filePath) && c.SecretsPath == "" {
		return "", fmt.Errorf("relative path provided for ADMIN_PASSWORD_FILE but SECRETS_PATH is not set")
	}

	hashBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read admin password file from %s: %w", filePath, err)
	}

	return strings.TrimSpace(string(hashBytes)), nil
}

// Returns the list of allowed CORS origins
func (c *Config) GetAllowedOrigins() []string {
	if c.AllowedOrigings == "" {
		// Default to localhost for dev instance
		return []string{"http://localhost:3000"}
	}

	// Split comma seperated origins and trim whitespace
	origins := strings.Split(c.AllowedOrigings, ",")
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
