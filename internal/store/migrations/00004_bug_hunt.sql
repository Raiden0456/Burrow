-- +goose Up
ALTER TABLE drills ADD COLUMN mode TEXT NOT NULL DEFAULT 'implement' CHECK (mode IN ('implement', 'bug_hunt'));
ALTER TABLE drills ADD COLUMN explanation TEXT NOT NULL DEFAULT '';
ALTER TABLE drills ADD COLUMN initial_output TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE drills DROP COLUMN initial_output;
ALTER TABLE drills DROP COLUMN explanation;
ALTER TABLE drills DROP COLUMN mode;
