package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/krosengr4/rosenblog-service/internal/config"
	models "github.com/krosengr4/rosenblog-service/internal/model"
	database "github.com/krosengr4/rosenblog-service/internal/repository"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	db     *database.DB
	config *config.Config
}

func New(db *database.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Writes a JSON response
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Error encoding the JSON response")
	}
}

// Writes an error response
func writeErrorResponse(w http.ResponseWriter, status int, messsage string) {
	log.Warn().Int("status:", status).Str("message:", messsage).Msg("Writing error response")
	writeJSONResponse(w, status, ErrorResponse{Error: messsage})
}

// GET /api/posts - Get all posts
func (h *Handler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /api/posts - Getting all posts")

	posts, err := h.db.GetAllPosts()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all posts")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Info().Int("count", len(posts)).Msg("Successfully got all posts")
	writeJSONResponse(w, http.StatusOK, posts)
}

// GET /api/posts/{slug} - Get a post by slug
func (h *Handler) GetPostBySlug(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /api/posts/{slug} - Getting post by slug")

	vars := mux.Vars(r)
	slug := vars["slug"]

	post, err := h.db.GetPostBySlug(slug)
	if err != nil {
		if err.Error() == "post not found" {
			log.Warn().Str("slug", slug).Msg("Post for that slug was not found")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		log.Error().Err(err).Str("slug", slug).Msg("Failed to get post by slug")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Info().Str("slug", slug).Msg("Successfully retrieved post by slug")
	writeJSONResponse(w, http.StatusOK, post)
}

// GET /api/posts/search?q=keyword - Handler for searching by title, content, or tags
func (h *Handler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /api/posts/search?q=keyword - Searching posts by title, content, or tags")

	query := r.URL.Query().Get("q")
	if query == "" {
		log.Warn().Msg("No Query parameter 'q' and it is required")
		writeErrorResponse(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	posts, err := h.db.SearchPosts(query)
	if err != nil {
		log.Error().Err(err).Str("query", query).Msg("Failed to search posts")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to search posts")
		return
	}

	log.Info().Str("query", query).Msg("Successfully searched and retrieved posts")
	writeJSONResponse(w, http.StatusOK, posts)
}

// GET /api/tags - Handler for getting all tags
func (h *Handler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /api/tags - Getting all tags")

	tags, err := h.db.GetAllTags()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tags")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get tags")
		return
	}

	log.Info().Int("count", len(tags)).Msg("Successfully got all tags")
	writeJSONResponse(w, http.StatusOK, tags)
}

// * Admin Handlers

// Slugify converts a title into a URL-safe slug
func slugify(title string) string {
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	slug = truncateSlug(slug, 40)
	return slug
}

// truncates the slug to a max length, breaking at a word boundary
func truncateSlug(slug string, maxLen int) string {
	if len(slug) <= maxLen {
		return slug
	}
	truncated := slug[:maxLen]
	if i := strings.LastIndex(truncated, "-"); i > 0 {
		truncated = truncated[:i]
	}
	return truncated
}

// POST /admin/posts - Handler to create a post
func (h *AdminHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /admin/posts - Creating new post")

	var req models.Post

	// Decode the JSON request body into CreatePost struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields before hitting the database
	if req.Title == "" || req.Content == "" {
		log.Warn().Msg("Missing required fields")
		writeErrorResponse(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Create Post object
	post := models.Post{
		Title:       req.Title,
		Slug:        slugify(req.Title),
		Content:     req.Content,
		Author:      "Kevin Rosengren",
		Tags:        req.Tags,
		PublishedAt: time.Now(),
	}

	// Insert Post into DB
	result, err := h.db.Create(&post)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	log.Info().Str("Title:", result.Title).Msg("Post created successfuly")
	writeJSONResponse(w, http.StatusOK, post)
}

// PUT /admin/posts/{postID} - Handler to update a post
func (h *AdminHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("PUT /admin/posts/{postID} - Updating post")

	// Get postID from URL params
	vars := mux.Vars(r)
	idStr := vars["postID"]

	// Convert string ID into an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("post_id", idStr).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get existing post from database
	existingPost, err := h.db.GetPostByID(id)
	if err != nil {
		if err.Error() == "post not found" {
			log.Warn().Int("post_id", id).Msg("Post not found")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}

		log.Error().Err(err).Msg("failed to get post")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get post")
		return
	}

	// Request body
	var req struct {
		Title   string   `json:"title"`
		Slug    string   `json:"slug"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update past object with new info
	existingPost.Title = req.Title
	existingPost.Slug = req.Slug
	existingPost.Content = req.Content
	existingPost.Tags = req.Tags

	// Call database to add updated post
	if err := h.db.UpdatePost(existingPost); err != nil {
		log.Error().Err(err).Msg("Failed to update post")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to update the post")
		return
	}

	// Success
	log.Info().Int("post_id", id).Str("Title", existingPost.Title).Msg("Post updated successfully")
	writeJSONResponse(w, http.StatusOK, existingPost)
}

// DELETE /admin/posts/{postID} - Delete post
func (h *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("DELETE /admin/posts/{postID} - Deleting post")

	// Get postID from URL params
	vars := mux.Vars(r)
	idStr := vars["postID"]

	// Convert string ID into int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("Post_ID", idStr).Msg("Invalid post id format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid post id format")
		return
	}

	// Delete the post from the DB
	if err := h.db.DeletePost(id); err != nil {
		if err.Error() == "post not found" {
			log.Warn().Int("post_id", id).Msg("Post not found for deletion")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		log.Error().Err(err).Msg("Failed to delete post")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	log.Info().Int("PostID", id).Msg("Post was deleted successfully")
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Post deleted successfully"})
}
