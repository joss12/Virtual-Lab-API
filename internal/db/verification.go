// Package db
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateEmailVerificationParams struct {
	UserID    pgtype.UUID
	Token     string
	ExpiresAt time.Time
}

func (q *Queries) CreateEmailVerification(ctx context.Context, arg CreateEmailVerificationParams) error {
	_, err := q.db.Exec(ctx, `
	INSERT INTO email_verification (user_id, token, expires_at)
	VALUES ($1, $2, $3)
	`, arg.UserID, arg.Token, arg.ExpiresAt)
	return err
}

func (q *Queries) GetEmailVerification(ctx context.Context, token string) (pgtype.UUID, error) {
	var userID pgtype.UUID
	err := q.db.QueryRow(ctx, `
	SELECT user_id FROM email_verifications
	WHERE token = $1 AND expires_at > NOW()
	LIMIT 1
	`, token).Scan(&userID)
	return userID, err
}

func (q *Queries) VerifyUserEmail(ctx context.Context, userID pgtype.UUID) error {
	_, err := q.db.Exec(ctx, `
	UPDATE users SET email_verified = TRUE WHERE id = $1
	`, userID)
	return err
}

func (q *Queries) DeleteEmailVerification(ctx context.Context, token string) error {
	_, err := q.db.Exec(ctx, `
		DELETE FROM email_verifications WHERE token = $1
	`, token)
	return err
}
