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
