package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/vlab-api/internal/db"
	"github.com/vlab-api/internal/handler"
	"github.com/vlab-api/internal/middleware"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("✓ database connected")

	q := db.NewPool(pool)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"vlab-api"}`))
	})

	r.Post("/auth/register", handler.Register(q))
	r.Post("/auth/login", handler.Login(q))
	r.Get("/leaderboard", handler.GetLeaderboard(q))

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Get("/quiz/scores", handler.GetScores(q))
		r.Post("/quiz/score", handler.SaveScore(q))
		r.Get("/progress", handler.GetProgress(q))
		r.Post("/progress/{component}", handler.UpdateProgress(q))
		r.Put("/auth/password", handler.ChangePassword(q))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("✓ server listening on :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
