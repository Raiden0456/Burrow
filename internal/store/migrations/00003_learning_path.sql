-- +goose Up
CREATE TABLE learned_topics (
    topic_slug TEXT PRIMARY KEY,
    learned_at INTEGER NOT NULL
);

ALTER TABLE quizzes ADD COLUMN checkpoint BOOLEAN NOT NULL DEFAULT 0;

CREATE TABLE jobs_new (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('lesson', 'quiz', 'drill', 'daily', 'checkpoint')),
    topic_slug TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('running', 'done', 'failed')),
    result_id INTEGER,
    error TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    finished_at INTEGER
);
INSERT INTO jobs_new (id, kind, topic_slug, status, result_id, error, created_at, finished_at)
SELECT id, kind, topic_slug, status, result_id, error, created_at, finished_at FROM jobs;
DROP TABLE jobs;
ALTER TABLE jobs_new RENAME TO jobs;
CREATE INDEX jobs_status_idx ON jobs (status, finished_at);

-- +goose Down
CREATE TABLE jobs_old (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('lesson', 'quiz', 'drill', 'daily')),
    topic_slug TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('running', 'done', 'failed')),
    result_id INTEGER,
    error TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    finished_at INTEGER
);
INSERT INTO jobs_old (id, kind, topic_slug, status, result_id, error, created_at, finished_at)
SELECT id, kind, topic_slug, status, result_id, error, created_at, finished_at FROM jobs WHERE kind != 'checkpoint';
DROP TABLE jobs;
ALTER TABLE jobs_old RENAME TO jobs;
CREATE INDEX jobs_status_idx ON jobs (status, finished_at);

ALTER TABLE quizzes DROP COLUMN checkpoint;
DROP TABLE learned_topics;
