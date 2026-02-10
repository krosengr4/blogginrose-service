package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/krosengr4/rosenblog-service/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	cfg       *config.Config
	jwtSecret string
}

func NewAdminHandler(cfg *config.Config, jwtSecret string) *AdminHandler {
	return &AdminHandler{cfg: cfg, jwtSecret: jwtSecret}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check username
	if req.Username != h.cfg.AdminUsername {
		log.Warn().Str("username", req.Username).Msg("Invalid username for login attempt")
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Read bcrypt hash from file
	hash, err := h.cfg.GetAdminPasswordHash()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read admin password hash")
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Check username and password
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		log.Warn().Str("username", req.Username).Msg("Login attempt failed")
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT (24 hour expiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		log.Error().Err(err).Msg("Failed to sign JWT")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info().Str("username", req.Username).Msg("Admin login successful")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{Token: tokenString})
}
