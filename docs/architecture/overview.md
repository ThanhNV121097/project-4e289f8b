# Architecture Overview

Project: "Task board for one person"

Shape: fullstack. Product ships one Next.js UI, one Go REST API, and one PostgreSQL database. Scope is one tasks module, one tasks table, one screen. No accounts, users, boards, memberships, assignments, comments, labels, search, notifications, file upload, browser-storage persistence, or activity log.

## Stack

- Frontend: Next.js 15 App Router, TypeScript, Tailwind CSS v3, ESLint.
- Backend: Go 1.22 HTTP server.
- Database: PostgreSQL 16 in local compose and managed Postgres in deployment.
- Runtime: Docker Compose from repository root.
- CI: GitHub Actions checks Go build/vet/test, frontend lint/build/test, compose syntax, and CSS token policy.

## Repository layout

```text
code/
  backend/
    cmd/api/main.go              # API process entrypoint
    migrations/                  # SQL migrations embedded into API binary
    go.mod
    .env.example
    Dockerfile                   # committed container contract, do not move service
  frontend/
    app/layout.tsx               # App Router root layout
    app/page.tsx                 # composition root only
    app/globals.css              # shared design tokens and base classes, frozen after scaffold
    package.json
    next.config.js
    tailwind.config.ts
    postcss.config.js
    tsconfig.json
    eslint.config.mjs
    .env.example
    Dockerfile                   # committed container contract, standalone Next runtime
docs/
  tasks/SRS.md
  architecture/overview.md
  architecture/erd.md            # later task owns database design detail
  architecture/services.md       # later task owns endpoint contract detail
```

## Component boundaries and data flow

1. Browser loads `code/frontend/app/page.tsx`.
2. Story components mounted later call REST API through `NEXT_PUBLIC_API_URL`.
3. Go API reads `DATABASE_URL`, applies embedded migrations, then serves HTTP on `PORT`, falling back to `APP_PORT`, then `8080`.
4. API persists tasks in PostgreSQL only. Browser storage is never source of task truth.
5. `/healthz` returns `200` only after migrations succeeded and `SELECT 1` against database succeeds.

Detailed table shape and endpoint contracts are deliberately deferred to ERD and service design tasks. Scaffold only proves boot, migration runner, DB connectivity, frontend build, and CI gate.

## Naming conventions

- Module slug: `tasks`.
- Frontend components: PascalCase filenames under `code/frontend/components/` when stories add them.
- Frontend component declaration: `export default function ComponentName()`.
- Client components: file first line must be literal `"use client"` when using event handlers, state, effects, refs, or browser APIs.
- `app/page.tsx`: server component composition root only. Later stories add one import and one JSX element.
- Backend packages: keep exactly one `main` package under `cmd/api` so `go build ./...` produces one binary.
- Migrations: timestamped `*.up.sql` and matching `*.down.sql`, applied in filename order, recorded in `schema_migrations`.
- Env vars: UPPER_SNAKE_CASE. Every read key appears in matching `.env.example` with comment and no secret value.

## Environment variables

### Root `.env.example`

- `POSTGRES_USER` — local Postgres user for Docker Compose.
- `POSTGRES_PASSWORD` — local Postgres password for Docker Compose.
- `POSTGRES_DB` — local Postgres database for Docker Compose.
- `BACKEND_PORT` — host port mapped to backend container port 8080.
- `FRONTEND_PORT` — host port mapped to frontend container port 3000.
- `NEXT_PUBLIC_API_URL` — browser-visible API base URL baked into Next.js bundle.
- `IMAGE_REPO` — optional image repository prefix used by compose.
- `IMAGE_TAG` — optional image tag used by compose.

### Backend `code/backend/.env.example`

- `DATABASE_URL` — PostgreSQL connection string injected by runtime.
- `PORT` — API listen port.
- `APP_PORT` — optional fallback listen port.

### Frontend `code/frontend/.env.example`

- `NEXT_PUBLIC_API_URL` — browser-visible REST API base URL.

## Design tokens

`code/frontend/app/globals.css` owns shared values from `design/design-system.md`: color, spacing, typography, radius, shadow, and motion. Story CSS modules must use `var(--token)` and must not include hardcoded hex colors or token fallbacks. CI enforces this for `*.module.css`.

## Security and reliability

- No authentication by design; API still treats all request bodies as untrusted external input.
- Later API handlers must validate title length, description length, status enum, date-only due date, and task ID at boundary.
- SQL must use parameterized queries.
- Health check must include database connectivity, not process liveness only.
- Migration runner is idempotent and locks migration rows through a single DB transaction per file.
- External errors returned to UI should be useful but not expose database internals.
- Backend uses request timeouts and graceful shutdown for process signals.

## Observability

- Backend logs startup, migration failures, listen address, and shutdown errors to stdout/stderr.
- Docker health checks cover API health and frontend reachability.
- CI publishes failing command logs through GitHub Actions.

## How to run

```bash
cp .env.example .env
docker compose --profile local up --build
```

Open frontend at `http://localhost:3000`. API listens on `http://localhost:8080`. Backend health check: `http://localhost:8080/healthz`.

Local service checks:

```bash
cd code/backend && go build ./... && go vet ./... && go test ./...
cd code/frontend && npm ci && npm run lint && npm run build && npm test --if-present
docker compose --profile local config -q
```

## Key decisions

### Fullstack shape

Decision: build frontend, backend, and database because stakeholder requires browser reload to show exact Postgres state behind REST API.

Rejected alternative: static frontend with browser storage. Tradeoff: lower build cost, but violates persistence requirement and acceptance criteria.

Rejected alternative: stateless backend without database. Tradeoff: simpler deployment, but cannot prove durable task state after reload.

### Go backend with stdlib HTTP

Decision: use Go 1.22, `net/http`, `database/sql`, embedded SQL migrations, and `github.com/jackc/pgx/v5/stdlib` as Postgres driver.

Rejected alternative: add web framework. Tradeoff: router helpers, but unnecessary for tiny REST surface and adds dependency surface.

Rejected alternative: separate migration tool. Tradeoff: mature migration CLI, but runtime database starts empty and self-migration on boot is required.

### Next.js App Router frontend

Decision: use Next.js 15 App Router with server `app/page.tsx` composition root and client components only where interaction requires them.

Rejected alternative: single client-only page. Tradeoff: quick UI implementation, but weaker boundary discipline and easier to break App Router build rules.

Rejected alternative: frontend-only mock data. Tradeoff: faster first screen, but risks story code persisting client fixtures after API exists.

### Tailwind plus CSS custom properties

Decision: Tailwind handles layout utilities; `globals.css` defines design-system tokens. Story CSS modules consume tokens.

Rejected alternative: hardcoded values per component. Tradeoff: local speed, but causes visual drift and repeat review failures.

Rejected alternative: CSS-in-JS library. Tradeoff: component-local styles, but adds dependency without need.

### Docker Compose as run contract

Decision: `docker compose --profile local up --build` boots Postgres, backend, and frontend together.

Rejected alternative: separate manual commands. Tradeoff: simpler files, but no single reproducible runtime for Dev, Test, and CI.

## Risks and unknowns

- ERD and service design still must define exact task table, indexes, and endpoint payloads.
- Due date is date-only; timezone-aware deadlines are new scope.
- No user boundary exists; deployment should be treated as single-person private tool unless product scope changes.
