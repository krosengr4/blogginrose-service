package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/krosengr4/rosenblog-service/internal/config"
	"github.com/krosengr4/rosenblog-service/internal/handler"
	"github.com/krosengr4/rosenblog-service/internal/middleware"
	database "github.com/krosengr4/rosenblog-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}).With().Timestamp().Logger()

	log.Info().Msg("Starting BlogginRose Backend Service!")

	// Load configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configurations")
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Resolve JWT secret from file once at startup
	jwtSecret, err := cfg.GetJWTSecret()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read JWT secret")
	}

	// Initialize handlers
	blogHandler := handler.New(db, cfg)
	adminHandler := handler.NewAdminHandler(cfg, db, jwtSecret)

	// Setup router
	router := setupRouter(blogHandler, adminHandler, jwtSecret)

	// Initialize CORS middleware with configuration
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.GetAllowedOrigins(),
	}

	// Apply middleware chain: Recover -> Logging -> CORS
	httpHandler := middleware.Recovery(
		middleware.Logging(
			middleware.CORS(corsConfig)(router),
		),
	)

	log.Info().Str("Port:", cfg.Port).Msg("BlogginRose service starting")

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal().Err(server.ListenAndServe()).Msg("Server failed to start")

}

// Setup router configures all of the API routes
func setupRouter(h *handler.Handler, adminHandler *handler.AdminHandler, jwtSecret string) *mux.Router {
	router := mux.NewRouter()

	// Set up API routes
	api := router.PathPrefix("/api").Subrouter()

	// Admin routes
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.JWTAuth(jwtSecret))

	// Get all posts
	api.HandleFunc("/posts", h.GetAllPosts).Methods("GET")
	// Search by content, title, or tags
	api.HandleFunc("/posts/search", h.SearchPosts).Methods("GET")
	// Get post by slug
	api.HandleFunc("/posts/{slug}", h.GetPostBySlug).Methods("GET")

	// Login for admin
	api.HandleFunc("/login", adminHandler.Login).Methods("POST")
	// Admin create post
	admin.HandleFunc("/posts", adminHandler.CreatePost).Methods("POST")

	return router
}
