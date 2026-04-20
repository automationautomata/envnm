-- name: UpsertVariables :exec
INSERT INTO variables (key, value, environment_id)
SELECT
  u.key,
  u.value,
  u.environment_id
FROM UNNEST($1::variable_entry[])
AS u(key, value, environment_id)
ON CONFLICT (key, environment_id)
DO UPDATE SET value = EXCLUDED.value;

-- name: GetVariablesByEnv :many
SELECT key, value FROM variables WHERE environment_id = $1;

-- name: DeleteVariable :exec
DELETE FROM variables WHERE environment_id = $1 AND key = $2;
