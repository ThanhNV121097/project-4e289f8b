# Technical Design — View tasks by status

This story uses existing `tasks` table and `GET /api/v1/tasks` contract already documented in `docs/architecture/erd.md` and `docs/architecture/services.md`.

Mock contract from approved UI PR matches service design:

- list envelope: `{ tasks, next_cursor, has_more }`
- task fields: `id`, `title`, `description`, `status`, `due_date`, `created_at`, `updated_at`
- nullable fields: `description`, `due_date`, `next_cursor`
- statuses: `todo`, `doing`, `done`
- error envelope: `{ error: { code, message, details, request_id } }`

Schema extension: none. Existing `tasks` table covers TASKS-001.

Endpoint extension: none. Existing `GET /api/v1/tasks` covers TASKS-001.

Migration plan: no new migration. Forward no-op, backward no-op, safe on populated tables because no database change occurs.
