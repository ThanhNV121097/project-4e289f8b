CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title text NOT NULL,
    description text NULL,
    status text NOT NULL DEFAULT 'todo',
    due_date date NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_tasks_title_length CHECK (length(title) BETWEEN 1 AND 120),
    CONSTRAINT ck_tasks_description_length CHECK (description IS NULL OR length(description) <= 2000),
    CONSTRAINT ck_tasks_status CHECK (status IN ('todo', 'doing', 'done'))
);

CREATE INDEX idx_tasks_status_created_at ON tasks (status, created_at);
