CREATE TABLE IF NOT EXISTS learn_progress (
    id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language    VARCHAR(50)  NOT NULL,
    lesson_slug VARCHAR(100) NOT NULL,
    completed   BOOLEAN      NOT NULL DEFAULT FALSE,
    quiz_score  INT,
    quiz_total  INT,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, language, lesson_slug)
);

CREATE INDEX idx_learn_progress_user_lang ON learn_progress (user_id, language);
