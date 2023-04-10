-- name: GetFact :one
SELECT id, created_at, updated_at, deleted_at, content, source
FROM facts
WHERE id = ? AND deleted_at IS NULL LIMIT 1;

-- name: GetFacts :many
SELECT id, created_at, updated_at, deleted_at, content, source
FROM facts
WHERE deleted_at IS NULL;

-- name: GetRandomFact :one
SELECT id, created_at, updated_at, deleted_at, content, source
FROM facts
WHERE deleted_at IS NULL
ORDER BY RANDOM() LIMIT 1;

-- name: CreateFact :one
INSERT INTO facts (content, source) VALUES (?, ?)
RETURNING id, created_at, updated_at, deleted_at, content, source;

-- name: DeleteFact :exec
DELETE FROM facts WHERE id = ?;

-- name: SoftDeleteFact :exec
UPDATE facts
SET deleted_at = DATETIME('now')
WHERE id = ?;
