CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Один ключ доступа может использоваться несколькими приложениями-клиентами
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

-- Окружение может иметь ключ доступа, но это не обязательно
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
