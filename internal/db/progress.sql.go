package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UpsertProgressParams struct {
	UserID      pgtype.UUID
	Component   string
	TabsVisited []string
	Completed   bool
}

func (q *Queries) UpsertProgress(ctx context.Context, arg UpsertProgressParams) (Progress, error) {
	row := q.db.QueryRow(ctx, `
	INSERT INTO progress (user_id, component, tabs_visited, completed)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, component) DO UPDATE SET
			tabs_visited = EXCLUDED.tabs_visited,
			completed    = EXCLUDED.completed,
			updated_at   = now()
		RETURNING id, user_id, component, tabs_visited, completed, updated_at
	`, arg.UserID, arg.Component, arg.TabsVisited, arg.Completed)

	var p Progress
	err := row.Scan(&p.ID, &p.UserID, &p.Component, &p.TabsVisited, &p.Completed, &p.UpdatedAt)
	return p, err
}

func (q *Queries) GetUserProgress(ctx context.Context, userID pgtype.UUID) ([]Progress, error) {
	rows, err := q.db.Query(ctx, `
	SELECT id, user_id, component, tabs_visited, completed, updated_at
		FROM progress
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Progress
	for rows.Next() {
		var p Progress
		if err := rows.Scan(&p.ID, &p.UserID, &p.Component, &p.TabsVisited, &p.Completed, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}

	if items == nil {
		items = []Progress{}
	}
	return items, nil
}
