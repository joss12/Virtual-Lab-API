package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type InsertQuizScoreParams struct {
	UserID    pgtype.UUID
	Score     int32
	Total     int32
	Component pgtype.Text
}

func (q *Queries) InsertQuizScore(ctx context.Context, arg InsertQuizScoreParams) (QuizScore, error) {
	row := q.db.QueryRow(ctx, `
		INSERT INTO quiz_scores (user_id, score, total, component)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, score, total, component, created_at
	`, arg.UserID, arg.Score, arg.Total, arg.Component)

	var s QuizScore
	err := row.Scan(&s.ID, &s.UserID, &s.Score, &s.Total, &s.Component, &s.CreatedAt)
	return s, err
}

func (q *Queries) GetUserScores(ctx context.Context, userID pgtype.UUID) ([]QuizScore, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, user_id, score, total, component, created_at
		FROM quiz_scores
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []QuizScore
	for rows.Next() {
		var s QuizScore
		if err := rows.Scan(&s.ID, &s.UserID, &s.Score, &s.Total, &s.Component, &s.CreatedAt); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}

	if scores == nil {
		scores = []QuizScore{}
	}
	return scores, nil
}

type LeaderboardEntry struct {
	Rank      int       `json:"rank"`
	Email     string    `json:"email"`
	Score     int32     `json:"score"`
	Total     int32     `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

func (q *Queries) GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := q.db.Query(ctx, `
    SELECT
        RANK() OVER (ORDER BY qs.score DESC, qs.created_at ASC) as rank,
        u.email,
        qs.score,
        qs.total,
        qs.created_at
    FROM quiz_scores qs
    JOIN users u ON u.id = qs.user_id
    ORDER BY qs.score DESC, qs.created_at ASC
    LIMIT 10
`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.Email, &e.Score, &e.Total, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []LeaderboardEntry{}
	}
	return entries, nil
}
