-- name: ListAnswerHistory :many
SELECT q.topic_slug, a.question_id, a.score, a.created_at
FROM answers a
JOIN questions q ON q.id = a.question_id
ORDER BY a.id;

-- name: CountSolvedDrillsByTopic :many
SELECT topic_slug, COUNT(*) AS solved
FROM drills
WHERE solved_at IS NOT NULL
GROUP BY topic_slug;

-- name: ListDrillSolveTimes :many
SELECT solved_at FROM drills WHERE solved_at IS NOT NULL;

-- name: ListCardStates :many
SELECT question_id, state FROM cards;

-- name: ListLearnedTopics :many
SELECT topic_slug FROM learned_topics;

-- name: MarkTopicLearned :exec
INSERT INTO learned_topics (topic_slug, learned_at) VALUES (?, ?)
ON CONFLICT (topic_slug) DO NOTHING;
