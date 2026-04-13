package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreatePasswordResetParams struct {
	UserID    pgtype.UUID
	Token     string
	ExpiresAt time.Time
}

func (q *Queries) CreatePasswordReset(ctx context.Context, arg CreatePasswordResetParams) error {
	_, err := q.db.Exec(ctx, `
	INSERT INTO password_resets (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, arg.UserID, arg.Token, arg.ExpiresAt)
	return err
}

func (q *Queries) GetPasswordReset(ctx context.Context, token string) (PasswordReset, error) {
	row := q.db.QueryRow(ctx, `
	SELECT id, user_id, token, expires_at, created_at
		FROM password_resets
		WHERE token = $1 AND expires_at > NOW()
		LIMIT 1
	`, token)
	var r PasswordReset
	err := row.Scan(&r.ID, &r.UserID, &r.Token, &r.ExpiresAt, &r.CreatedAt)
	return r, err
}

func (q *Queries) DeletePasswordReset(ctx context.Context, token string) error {
	_, err := q.db.Exec(ctx, `
	DELETE FROM password_resets WHERE token = $1
	`, token)
	return err
}

func (q *Queries) DeleteExpiredPasswordResets(ctx context.Context) error {
	_, err := q.db.Exec(ctx, `
	DELETE FROM password_resets WHERE expires_at < NOW()
	`)
	return err
}
