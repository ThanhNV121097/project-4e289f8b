# Story — Create task

Module: `tasks`
Plan item: Create task
Requirement: TASKS-002 — Create persisted task

## User story

As a Board owner, I want to create a task with title, optional description, due date, and initial status, so that new work is saved in Postgres and survives browser reload.

## In scope

- Create-task UI on the approved one-screen task board.
- Title input, required, trimmed, 1 to 120 characters.
- Optional description input, trimmed, up to 2,000 characters; blank means absent.
- Optional date-only due date input; blank means absent.
- Initial status select with `todo`, `doing`, and `done`; default is `todo`.
- Client-side validation before submit for required title, field length limits, date validity, and allowed status values.
- REST API create call that persists through Postgres.
- Board update from the saved task returned by the API, not from browser storage.
- Browser reload showing the created task from API-backed database state.
- Inline validation and save-failure messaging.
- Toast or equivalent short confirmation after successful create, using approved design copy pattern.

## Out of scope

- Sign-up, sign-in, accounts, sessions, users, roles, and permissions beyond Board owner.
- Multiple users, members, assignments, sharing, multiple boards, or multiple projects.
- Comments, labels, tags, activity log, file upload, search, notifications, and archived boards.
- Browser storage persistence through localStorage, sessionStorage, IndexedDB, cookies, URL parameters, or hard-coded fixtures.
- Creating more than one entity type.
- Bulk create, import, templates, recurrence, priorities, reminders, time-of-day deadlines, and timezone-aware deadlines.
- Drag-and-drop movement; movement belongs to Edit and move task.
- Delete, edit existing task details, and move existing tasks after creation.
- Mockup-only “Simulate reload from API” button and blue “Persistence rule” banner.

## UI scope

This story touches the approved Task board screen only.

- Use `CreateTaskPanel` at the bottom create area with title, description, due date, status, and submit controls.
- Use `Field` patterns for text input, textarea, native date input, and native select.
- Use primary `Button` for create submit.
- Use inline `Field` invalid states for validation errors.
- Update `BoardColumn`, `TaskCard`, `StatusPill`, `StateSummaryCard`, `EmptyState`, and `Toast` only as needed to reflect newly saved API result.
- Preserve approved responsive behaviour: single column below `900px`, three board columns at `900px` and above, no horizontal page scroll from `320px` upward.
- Preserve approved motion: no page-load animation; hover/status-change transitions under 200ms; add reduced-motion handling during implementation where approved design lacked it.
- Keep keyboard path to every create control and visible focus indicators.

## Acceptance criteria

1. Given create form is empty, when Board owner enters title `Pay invoice` and submits without changing status, then API persists task and board shows `Pay invoice` in Todo column from saved API response.
2. Given created task appears on board, when Board owner reloads browser, then same task appears from REST API data backed by Postgres.
3. Given Board owner enters description `Due before Friday`, when task is created, then created card shows `Due before Friday`.
4. Given Board owner chooses due date, when task is created, then created card shows that date as a date-only value without time.
5. Given Board owner chooses initial status `doing`, when task is created, then created card appears in Doing column and Doing count increases.
6. Given Board owner chooses initial status `done`, when task is created, then created card appears in Done column and Done count increases.
7. Given Board owner leaves description and due date blank, when titled task is created, then task is saved with absent description and absent due date, and card does not require either value.
8. Given title is blank after trimming whitespace, when Board owner submits, then inline title error appears and no create API request saves a task.
9. Given title is 120 trimmed characters, when Board owner submits valid form, then task is accepted and saved.
10. Given title is 121 trimmed characters, when Board owner submits, then inline title error names 120 character limit and task is not saved.
11. Given description is 2,000 trimmed characters, when Board owner submits valid form, then task is accepted and saved.
12. Given description is 2,001 trimmed characters, when Board owner submits, then inline description error names 2,000 character limit and task is not saved.
13. Given due date control contains invalid calendar date, when Board owner submits, then inline due date error appears and task is not saved.
14. Given status value is not `todo`, `doing`, or `done`, when create is submitted, then status error appears and task is not saved.
15. Given REST API or Postgres save fails, when Board owner submits valid form, then error message appears and board does not add task as persisted data.
16. Given another task already has same title, when Board owner creates task with duplicate title, then task is allowed and saved because titles are not unique.
17. Given create succeeds, when board updates, then totals and matching status column update from saved API response, not optimistic browser-only state.
18. Given browser storage panels are inspected after create, then task persistence is not present in localStorage, sessionStorage, IndexedDB, or cookies.

## Dependencies

- Depends on approved task board design and `design/design-system.md`.
- Depends on `docs/architecture/overview.md` constraints: Next.js frontend, Go REST API, PostgreSQL, Docker Compose run contract.
- Depends on REST API create endpoint and Postgres tasks table being defined by architecture/service design stages.
- Depends on View tasks by status for initial board loading and rendering persisted task columns.
- No external service accounts or credentials are needed.

## Non-blocking product decisions

- Default initial status is `todo`.
- Duplicate titles are allowed.
- Due date remains date-only, no time-of-day or timezone controls.
- Last saved API response is source of truth for immediate board update.
