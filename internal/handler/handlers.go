package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

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

// * Admin Handlers

// Slugify converts a title into a URL-safe slug
func slugify(title string) string {
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
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
