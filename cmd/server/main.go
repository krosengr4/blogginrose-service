package main

import (
	"os"

	"github.com/krosengr4/rosenblog-service/internal/config"
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

}
