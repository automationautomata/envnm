-- name: GetPoliciesByEnv :many
SELECT p.id, p.name, p.key, ep.changes_allowed
FROM environments_access_policies ep
JOIN access_policies p ON p.id = ep.access_policy_id
WHERE ep.environment_id = $1;

-- name: FindPolicyByID :one
SELECT id, name, key 
FROM access_policies 
WHERE id = $1;

-- name: FindPolicyByName :one
SELECT id, name, key 
FROM access_policies 
WHERE name = $1;

-- name: FindPolicyByKey :one
SELECT id, name, key 
FROM access_policies 
WHERE key = $1;

-- name: CreatePolicy :exec
INSERT INTO access_policies (id, name, key)
VALUES ($1, $2, $3);

-- name: AddPolicyToEnvironment :exec
INSERT INTO environments_access_policies (environment_id, access_policy_id, changes_allowed)
VALUES ($1, $2, $3);

-- name: DeletePolicyFromEnvironment :exec
DELETE FROM environments_access_policies
WHERE environment_id = $1 AND access_policy_id = $2;

-- name: DeletePolicy :exec
DELETE FROM access_policies
WHERE id = $1;
