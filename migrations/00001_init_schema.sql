-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE quiz_scores (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score       INT         NOT NULL CHECK (score >= 0),
    total       INT         NOT NULL CHECK (total > 0),
    component   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE progress (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    component    TEXT        NOT NULL,
    tabs_visited TEXT[]      NOT NULL DEFAULT '{}',
    completed    BOOLEAN     NOT NULL DEFAULT false,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, component)
);

CREATE INDEX idx_quiz_scores_user ON quiz_scores(user_id);
CREATE INDEX idx_progress_user    ON progress(user_id);


-- +goose Down
DROP TABLE IF EXISTS progress;
DROP TABLE IF EXISTS quiz_scores;
DROP TABLE IF EXISTS users;
