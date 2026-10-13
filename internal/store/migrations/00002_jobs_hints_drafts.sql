-- +goose Up
CREATE TABLE jobs (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('lesson', 'quiz', 'drill', 'daily')),
    topic_slug TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('running', 'done', 'failed')),
    result_id INTEGER,
    error TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    finished_at INTEGER
);
CREATE INDEX jobs_status_idx ON jobs (status, finished_at);

CREATE TABLE hints (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('hint', 'explain', 'drill')),
    quiz_id INTEGER REFERENCES quizzes (id) ON DELETE CASCADE,
    question_id INTEGER REFERENCES questions (id) ON DELETE CASCADE,
    drill_id INTEGER REFERENCES drills (id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX hints_quiz_idx ON hints (quiz_id, question_id);
CREATE INDEX hints_drill_idx ON hints (drill_id);

ALTER TABLE quiz_items ADD COLUMN draft TEXT NOT NULL DEFAULT '';
ALTER TABLE drills ADD COLUMN draft TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE drills DROP COLUMN draft;
ALTER TABLE quiz_items DROP COLUMN draft;
DROP TABLE hints;
DROP TABLE jobs;
