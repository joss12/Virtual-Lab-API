// Package db
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID           pgtype.UUID `json:"id"`
	Email        string      `json:"email"`
	PasswordHash string      `json:"-"`
	GithubID     pgtype.Text `json:"github_id"`
	CreatedAt    time.Time   `json:"created_at"`
}

type QuizScore struct {
	ID        pgtype.UUID `json:"id"`
	UserID    pgtype.UUID `json:"user_id"`
	Score     int32       `json:"score"`
	Total     int32       `json:"total"`
	Component pgtype.Text `json:"component"`
	CreatedAt time.Time   `json:"created_at"`
}

type Progress struct {
	ID          pgtype.UUID `json:"id"`
	UserID      pgtype.UUID `json:"user_id"`
	Component   string      `json:"component"`
	TabsVisited []string    `json:"tabs_visited"`
	Completed   bool        `json:"completed"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
