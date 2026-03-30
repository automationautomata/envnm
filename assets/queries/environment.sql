-- name: CreateEnvironment :exec
INSERT INTO environments (id, name, description, last_variables_update, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEnvironmentByID :one
SELECT * FROM environments WHERE id = $1;

-- name: GetEnvironmentByName :one
SELECT * FROM environments WHERE name = $1;

-- name: ListEnvironments :many
SELECT * FROM environments;

-- name: DeleteEnvironment :exec
DELETE FROM environments WHERE id = $1;
