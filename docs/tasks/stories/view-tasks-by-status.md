# Story — View tasks by status

Module: `tasks`
Plan item: View tasks by status
Requirement: TASKS-001 — Load persisted tasks into three status columns

## User story

As a Board owner, I want to see every persisted task grouped by status, so that I know what is todo, doing, and done after any browser reload.

## In scope

- Load tasks from REST API backed by Postgres when the task board page opens.
- Render exactly three columns: Todo, Doing, Done.
- Place every returned task in the column matching saved status: `todo`, `doing`, or `done`.
- Show task title, optional description when present, due date when present, current status label, and visible task actions on each card.
- Show status counts for Todo, Doing, Done from loaded task data.
- Show quiet empty state inside any status column with zero tasks.
- Show API loading state before task cards render.
- Show API error state with retry when task list request fails or returned data contains invalid status.
- Use REST API data as source of truth; do not use browser storage or hard-coded client fixtures as persisted board state.

## Out of scope

- Creating tasks.
- Editing task title, description, due date, or status.
- Moving tasks between statuses.
- Deleting tasks.
- Drag and drop.
- Search, filters, sorting controls, pagination, labels, tags, comments, notifications, activity log, file upload.
- Sign-up, sign-in, accounts, sessions, users, members, assignments, permissions, multiple boards, multiple projects.
- Browser-storage persistence through localStorage, sessionStorage, IndexedDB, cookies, URL parameters, or client fixtures.
- Mockup-only “Simulate reload from API” button and blue “Persistence rule” banner.

## UI scope

Touches one approved Task board screen.

Required UI elements for this story:

- Topbar and hero remain visible per approved design.
- State summary cards show Todo, Doing, Done counts.
- LoadingState appears during initial API fetch and retry fetch, with copy that says tasks load from API.
- ErrorState appears when tasks cannot load, with retry action and no stale browser-storage substitute.
- BoardColumn renders exactly three responsive columns: Todo, Doing, Done; columns stack below 900px and render side by side at 900px and above.
- TaskCard renders title, optional description, optional due date, status pill, and visible placeholders for later edit/move/delete actions if mounted by later stories.
- EmptyState renders `No todo tasks.`, `No doing tasks.`, or `No done tasks.` inside empty columns.

Design constraints:

- Use approved tokens from `design/design-system.md` through `globals.css` and component CSS module variables.
- Use neutral grey/slate surfaces, blue primary action, amber Doing tint, green Done tint.
- No gradients, no heavy color blocks, no page-load animation.
- Hover and status-change transitions stay under 200ms.
- Keyboard focus remains visible on retry and any rendered actions.

## Acceptance criteria

1. Given Postgres contains one `todo`, one `doing`, and one `done` task, when Board owner opens page, then each task appears in matching status column.
2. Given Postgres contains three tasks, when Board owner reloads browser, then page shows those same three tasks from REST API.
3. Given Postgres contains no `doing` tasks, when Board owner opens page, then Doing column shows quiet empty state `No doing tasks.`
4. Given Postgres contains one task, when Board owner opens browser storage panels, then no task persistence exists in localStorage, sessionStorage, IndexedDB, or cookies.
5. Given Postgres contains task with description and due date, when Board owner opens page, then card shows title, description, due date, and status label.
6. Given Postgres contains task without description and without due date, when Board owner opens page, then card shows title and status label without requiring description or due date text.
7. Given task list request fails or times out, when Board owner opens page, then API error state appears with retry action and no stale task cards from browser storage.
8. Given API returns empty list, when Board owner opens page, then all three columns render and each column shows quiet empty state.
9. Given API returns task with status outside `todo`, `doing`, `done`, when Board owner opens page, then board does not render a fourth column and shows API error state.
10. Given 200 tasks are returned by API, when response starts, then initial board render completes within 2 seconds in local development.
11. Given Board owner navigates by keyboard, when focus reaches retry and task actions, then visible focus indicator remains present.
12. Given viewport width is 320px, when Board owner opens page, then one-screen board has no horizontal page scroll.

## Dependencies

- Approved SRS: `docs/tasks/SRS.md` requirement TASKS-001.
- Approved design system: `design/design-system.md`.
- Architecture overview: `docs/architecture/overview.md`.
- REST API endpoint for listing tasks, backed by Postgres. Exact contract comes from TL service design.
- Postgres tasks table. Exact schema comes from TL ERD.
- No external accounts or credentials.

## Decisions

- Default task ordering inside each status column is API order; no sorting control in this story.
- Invalid persisted status is treated as API/data error, not hidden fourth status.
- Empty columns always stay visible to preserve three-column board shape.
