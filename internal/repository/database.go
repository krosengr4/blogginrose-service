package database

import (
	"database/sql"
	"fmt"

	"github.com/krosengr4/rosenblog-service/internal/config"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type DB struct {
	*sql.DB
}

// Create new database connection
func New(cfg *config.Config) (*DB, error) {
	// Get the DB URL
	databaseURL, err := cfg.GetDatabaseURL()
	if err != nil {
		return nil, fmt.Errorf("could not get database url: %w", err)
	}

	// Open connection to the databse
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection with the database: %w", err)
	}

	// Ping database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the databse: %w", err)
	}

	log.Info().Msg("Database succesfully connected!")
	return &DB{DB: db}, nil
}

// Get all posts - GET /api/posts
