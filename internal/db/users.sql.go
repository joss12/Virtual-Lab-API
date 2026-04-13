// Package db
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserParams struct {
	Email        string
	PasswordHash string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, github_id, created_at
	`, arg.Email, arg.PasswordHash)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.GithubID, &u.CreatedAt)
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, `
		SELECT id, email, password_hash, github_id, created_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`, email)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.GithubID, &u.CreatedAt)
	return u, err
}

func (q *Queries) GetUserByID(ctx context.Context, id pgtype.UUID) (User, error) {
	row := q.db.QueryRow(ctx, `
		SELECT id, email, password_hash, github_id, created_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`, id)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.GithubID, &u.CreatedAt)
	return u, err
}

func (q *Queries) UpdateUserPassword(ctx context.Context, id pgtype.UUID, passwordHash string) error {
	_, err := q.db.Exec(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, passwordHash, id)
	return err
}

func (q *Queries) GetUserByGithubID(ctx context.Context, githubID string) (User, error) {
	row := q.db.QueryRow(ctx,
		`
		SELECT id, email, password_hash, github_id, created_at
		FROM users
		WHERE github_id = $1
		LIMIT 1
	`, githubID)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.GithubID, &u.CreatedAt)
	return u, err
}

type CreateGithubUserParams struct {
	Email    string
	GithubID string
}

func (q *Queries) CreateGithubUser(ctx context.Context, arg CreateGithubUserParams) (User, error) {
	row := q.db.QueryRow(ctx, `
		INSERT INTO users (email, github_id, password_hash)
		VALUES ($1, $2, '')
		RETURNING id, email, password_hash, github_id, created_at
	`, arg.Email, arg.GithubID)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.GithubID, &u.CreatedAt)
	return u, err
}
