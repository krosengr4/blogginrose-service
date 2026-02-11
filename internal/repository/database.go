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

// GET /api/posts/id/{postID} - Get a post by its ID
func (db *DB) GetPostByID(postId int) (*models.Post, error) {
	query := "SELECT * FROM posts WHERE post_id = $1;"

	var post models.Post
	err := db.DB.QueryRow(query, postId).Scan(
		&post.PostId, &post.Title, &post.Slug, &post.Content,
		&post.Tags, &post.Author, &post.PublishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return &post, nil
}

// GET /api/posts/{slug} - Get a single post by its slug
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
func (db *DB) GetAllTags() ([]string, error) {
	query := `
		SELECT DISTINCT UNNEST(tags) as tag FROM posts
		ORDER BY tag
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query tags")
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// * Admin functions
// POST /admin/posts - Create a new post
func (db *DB) Create(post *models.Post) (models.Post, error) {
	insertQuery := `
		INSERT INTO posts (title, slug, content, tags, author, published_at)
		VALUES ($1, '', $2, $3, $4, $5)
		RETURNING post_id
	`

	err := db.QueryRow(insertQuery, post.Title, post.Content, post.Tags, post.Author, post.PublishedAt).Scan(&post.PostId)
	if err != nil {
		return models.Post{}, fmt.Errorf("failed to create post: %w", err)
	}

	// Generate slug with post_id suffix and update the row
	post.Slug = fmt.Sprintf("%s-%d", post.Slug, post.PostId)

	_, err = db.Exec("UPDATE posts SET slug = $1 WHERE post_id = $2", post.Slug, post.PostId)
	if err != nil {
		return models.Post{}, fmt.Errorf("failed to update slug: %w", err)
	}

	return *post, nil
}

// PUT /admin/posts/{postID} - Edit a post
func (db *DB) UpdatePost(post *models.Post) error {
	query := `
		UPDATE posts
		SET title = $2, slug = $3, content = $4, tags = $5, author = $6, published_at = $7 
		WHERE post_id = $1
	`

	result, err := db.Exec(
		query, post.PostId, post.Title, post.Slug,
		post.Content, post.Tags, post.Author, post.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Int("post_id", post.PostId).Msg("No rows affected, post not found or no changes made")
	}

	log.Info().Int("post_id", post.PostId).Int64("rows affected", rowsAffected).Msg("Post update query executed")

	log.Info().Int("post_id", post.PostId).Msg("Successfully updated post in database")
	return nil
}

// DELETE /admin/posts/{postID} - Delete a post
func (db *DB) DeletePost(postID int) error {
	query := "DELETE FROM posts WHERE post_id = $1"

	result, err := db.Exec(query, postID)
	if err != nil {
		log.Error().Err(err).Int("post_id", postID).Msg("Failed to execute post deletion query")
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Int("post_id", postID).Msg("No rows affected - post not found")
	}

	log.Info().Int("post_id", postID).Int64("rows affected", rowsAffected).Msg("Post deletion query executed")

	log.Info().Int("post_id", postID).Msg("Successfully deleted post from the database")
	return nil
}
