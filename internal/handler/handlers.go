package handler

import (
	"encoding/json"
	"net/http"

	"github.com/krosengr4/rosenblog-service/internal/config"
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
