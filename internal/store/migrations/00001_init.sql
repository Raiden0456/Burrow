-- +goose Up
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    expires_at INTEGER NOT NULL
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE providers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL CHECK (kind IN ('anthropic', 'openai', 'acp')),
    config TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE lessons (
    topic_slug TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE questions (
    id INTEGER PRIMARY KEY,
    topic_slug TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('single_choice', 'multi_choice', 'open')),
    prompt TEXT NOT NULL,
    options TEXT NOT NULL DEFAULT '[]',
    correct TEXT NOT NULL DEFAULT '[]',
    reference_answer TEXT NOT NULL DEFAULT '',
    explanation TEXT NOT NULL,
    difficulty INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX questions_topic_idx ON questions (topic_slug);

CREATE TABLE cards (
    question_id INTEGER PRIMARY KEY REFERENCES questions (id) ON DELETE CASCADE,
    due INTEGER NOT NULL,
    state TEXT NOT NULL
);
CREATE INDEX cards_due_idx ON cards (due);

CREATE TABLE quizzes (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('topic', 'review', 'daily')),
    topic_slug TEXT,
    daily_date TEXT UNIQUE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    completed_at INTEGER
);

CREATE TABLE quiz_items (
    quiz_id INTEGER NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    question_id INTEGER NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    PRIMARY KEY (quiz_id, position)
);

CREATE TABLE answers (
    id INTEGER PRIMARY KEY,
    question_id INTEGER NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    quiz_id INTEGER REFERENCES quizzes (id) ON DELETE SET NULL,
    response TEXT NOT NULL,
    score REAL NOT NULL,
    feedback TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX answers_question_idx ON answers (question_id);
CREATE INDEX answers_quiz_idx ON answers (quiz_id);

CREATE TABLE drills (
    id INTEGER PRIMARY KEY,
    topic_slug TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starter_code TEXT NOT NULL,
    test_code TEXT NOT NULL,
    solution TEXT NOT NULL,
    difficulty INTEGER NOT NULL,
    solved_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX drills_topic_idx ON drills (topic_slug);

CREATE TABLE drill_attempts (
    id INTEGER PRIMARY KEY,
    drill_id INTEGER NOT NULL REFERENCES drills (id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    passed INTEGER NOT NULL,
    output TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX drill_attempts_drill_idx ON drill_attempts (drill_id);

CREATE TABLE chat_threads (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    context TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES chat_threads (id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX chat_messages_thread_idx ON chat_messages (thread_id);

CREATE TABLE push_subscriptions (
    endpoint TEXT PRIMARY KEY,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- +goose Down
DROP TABLE push_subscriptions;
DROP TABLE chat_messages;
DROP TABLE chat_threads;
DROP TABLE drill_attempts;
DROP TABLE drills;
DROP TABLE answers;
DROP TABLE quiz_items;
DROP TABLE quizzes;
DROP TABLE cards;
DROP TABLE questions;
DROP TABLE lessons;
DROP TABLE providers;
DROP TABLE settings;
DROP TABLE sessions;
