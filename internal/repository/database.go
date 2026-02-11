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

// GET /api/posts - Get all posts
func (db *DB) GetAllPosts() ([]models.Post, error) {
	query := "SELECT * FROM posts ORDER BY published_at DESC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	var postList []models.Post
	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.PostId, &post.Title, &post.Slug, &post.Content, &post.Tags, &post.Author, &post.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		postList = append(postList, post)
	}

	return postList, nil
}

// Get /api/posts/{slug} - Get a single post by its slug
func (db *DB) GetPostBySlug(slug string) (*models.Post, error) {
	query := `
		SELECT * FROM posts
		WHERE slug = $1
	`

	var post models.Post
	err := db.DB.QueryRow(query, slug).Scan(
		&post.PostId, &post.Title, &post.Slug, &post.Content,
		&post.Tags, &post.Author, &post.PublishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return &post, nil
}

// GET /api/posts/search?q=keyword - Searches posts by title, content, or tags
func (db *DB) SearchPosts(search string) ([]models.Post, error) {
	query := `
		SELECT * FROM posts 
		WHERE (
			title ILIKE '%' || $1 || '%' 
			OR content ILIKE '%' || $1 || '%' 
			OR $1 = ANY(tags)
		)
		ORDER BY published_at DESC 
		LIMIT 20
	`

	rows, err := db.Query(query, search)
	if err != nil {
		log.Error().Err(err).Str("search", search).Msg("Failed to search posts")
		return nil, fmt.Errorf("failed to search for posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.PostId, &post.Title, &post.Slug, &post.Content,
			&post.Tags, &post.Author, &post.PublishedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan post in search")
			return nil, fmt.Errorf("failed to scan rows in search: %w", err)
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// GET /api/tags - Retrieves all unique tags used accross posts

// * Admin functions
// POST /admin/posts - Create a new post
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

// PUT /admin/posts/{postID} - Edit a post

// DELETE /admin/posts/{postID} - Delete a post
