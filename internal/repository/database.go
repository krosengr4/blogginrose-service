package database

import (
	"database/sql"
	"fmt"

	"github.com/krosengr4/rosenblog-service/internal/config"
	models "github.com/krosengr4/rosenblog-service/internal/model"
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

/*
- Get all posts
- Get post by slug
- Post post
- Edit post
- Delete post
*/

// * Admin functions
// POST /admin/posts
func (db *DB) Create(post *models.Post) (models.Post, error) {
	query := `
		INSERT INTO posts (title, slug, content, tags, author, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING post_id
	`

	err := db.QueryRow(query, post.Title, post.Slug, post.Content, post.Tags, post.Author, post.PublishedAt).Scan(&post.PostId)
	if err != nil {
		return models.Post{}, fmt.Errorf("failed to create post: %w", err)
	}

	return *post, nil
}
