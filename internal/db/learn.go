package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type LearnProgress struct {
	ID         pgtype.UUID `json:"id"`
	UserID     pgtype.UUID `json:"user_id"`
	Language   string      `json:"language"`
	LessonSlug string      `json:"lesson_slug"`
	Completed  bool        `json:"completed"`
	QuizScore  *int        `json:"quiz_score"`
	QuizTotal  *int        `json:"quiz_total"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (q *Queries) GetLearnProgress(ctx context.Context, userID pgtype.UUID, language string) ([]LearnProgress, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, user_id, language, lesson_slug, completed, quiz_score, quiz_total, updated_at
		FROM learn_progress
		WHERE user_id = $1 AND language = $2
	`, userID, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []LearnProgress
	for rows.Next() {
		var p LearnProgress
		if err := rows.Scan(&p.ID, &p.UserID, &p.Language, &p.LessonSlug, &p.Completed, &p.QuizScore, &p.QuizTotal, &p.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}
	if progress == nil {
		progress = []LearnProgress{}
	}
	return progress, nil
}

func (q *Queries) UpsertLearnProgress(ctx context.Context, userID pgtype.UUID, language string, lessonSlug string, completed bool, quizScore *int, quizTotal *int) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO learn_progress (user_id, language, lesson_slug, completed, quiz_score, quiz_total, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, language, lesson_slug)
		DO UPDATE SET completed = $4, quiz_score = $5, quiz_total = $6, updated_at = NOW()
	`, userID, language, lessonSlug, completed, quizScore, quizTotal)
	return err
}

type LearnProgressSummary struct {
	Language  string `json:"language"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

func (q *Queries) GetLearnProgressSummary(ctx context.Context, userID pgtype.UUID) ([]LearnProgressSummary, error) {
	rows, err := q.db.Query(ctx, `
		SELECT language, COUNT(*) FILTER (WHERE completed = true) AS completed, COUNT(*) AS total
		FROM learn_progress
		WHERE user_id = $1
		GROUP BY language
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []LearnProgressSummary
	for rows.Next() {
		var s LearnProgressSummary
		if err := rows.Scan(&s.Language, &s.Completed, &s.Total); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	if summaries == nil {
		summaries = []LearnProgressSummary{}
	}
	return summaries, nil
}
