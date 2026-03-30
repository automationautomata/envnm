-- name: UpsertVariable :exec
INSERT INTO variables (key, value, environment_id)
VALUES ($1, $2, $3)
ON CONFLICT (key, environment_id)
DO UPDATE SET value = EXCLUDED.value;

-- name: GetVariablesByEnv :many
SELECT key, value FROM variables WHERE environment_id = $1;

-- name: DeleteVariable :exec
DELETE FROM variables WHERE environment_id = $1 AND key = $2;
