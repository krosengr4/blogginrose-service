package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// Recovers from panics and logs them
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with the stack trace
				log.Error().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Interface("panic", err).
					Str("stack", string(debug.Stack())).
					Msg("Panic recovered")

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error": "Internal server error"}`)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
