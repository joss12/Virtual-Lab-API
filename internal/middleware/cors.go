// Package middleware
package middleware

import (
	chiCors "github.com/go-chi/cors"
	"net/http"
	"os"
	"strings"
)

func CORS() func(http.Handler) http.Handler {
	origins := []string{"http://localhost:3000"}

	if extra := os.Getenv("FRONTEND_URL"); extra != "" {
		for o := range strings.SplitSeq(extra, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	}

	return chiCors.Handler(chiCors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
