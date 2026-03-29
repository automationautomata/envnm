CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Один ключ доступа может использоваться несколькими приложениями-клиентами
CREATE TABLE access_policies (
    id uuid PRIMARY KEY,
    name varchar(255) UNIQUE NOT NULL,
    key text NOT NULL UNIQUE,
    is_changes_allowed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE variables (
    id serial PRIMARY KEY,
    name varchar(255) UNIQUE NOT NULL,
    value text NOT NULL,
    is_deleted boolean NOT NULL DEFAULT false,
    old_variable_id integer REFERENCES variables(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE environments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL UNIQUE,
    description text,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Окружение может иметь ключ доступа, но это не обязательно
CREATE TABLE environments_access_keys (
    access_key_id serial NOT NULL REFERENCES access_keys(id) ON DELETE SET NULL,
    environment_id uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    UNIQUE (access_keys_id, environment_id)
);


