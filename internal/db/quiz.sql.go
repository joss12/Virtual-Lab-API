package db

import (
	"context"

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
