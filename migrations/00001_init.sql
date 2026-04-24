-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- One access policy key can be used by multiple client applications
CREATE TABLE access_policies (
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    key text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE environments (
    id uuid PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE,
    description text,
    last_variables_update timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE variables (
    id serial PRIMARY KEY,
    key varchar(255) NOT NULL,
    value text NOT NULL,
    environment_id uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (key, environment_id)
);

-- An environment may have an access policy, but it is optional
CREATE TABLE environments_access_policies (
    access_policy_id uuid NOT NULL REFERENCES access_policies(id) ON DELETE CASCADE,
    environment_id uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    changes_permission boolean NOT NULL DEFAULT false,
    UNIQUE (access_policy_id, environment_id)
);

CREATE TYPE variable_entry AS (
    key varchar(255),
    value text,
    environment_id uuid
);

-- +goose Down
DROP TABLE IF EXISTS environments_access_policies;
DROP TABLE IF EXISTS variables;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS access_policies;
DROP TYPE IF EXISTS variable_entry;