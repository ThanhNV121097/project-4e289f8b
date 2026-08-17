# Story — Edit and move task

Module: `tasks`
Plan item: Edit and move task
Requirements: TASKS-003, TASKS-004

## User story

As a Board owner, I want to edit a task title, optional description, due date, and status, so that saved task details and workflow position stay accurate after browser reload.

## In scope

- Open edit controls for one existing persisted task from its task card.
- Show current saved title, description, due date, and status in editable controls.
- Edit title, optional description, optional due date, and status.
- Validate edited fields before saving.
- Save edits through the REST API to Postgres.
- Move a task between `todo`, `doing`, and `done` through card move controls or edit controls.
- Update the task card, target status column, and status totals from the API response after save.
- Preserve saved edits and moved status after browser reload by reloading task state from the REST API.
- Show inline validation errors for editable fields.
- Show useful not-found, conflict, and upstream failure messages without using browser storage as fallback.

## Out of scope

- Creating tasks; covered by Create task.
- Deleting tasks; covered by Delete task.
- Viewing initial persisted tasks beyond what this story needs to edit or move; covered by View tasks by status.
- Sign-up, sign-in, accounts, users, roles, permissions, members, assignments, comments, labels, tags, activity log, file upload, search, notifications, multiple boards, and multiple projects.
- Drag-and-drop movement; keyboard-reachable move controls are enough.
- Statuses outside `todo`, `doing`, and `done`.
- Due time, timezone-aware deadlines, reminders, or recurrence.
- Browser storage persistence, optimistic persistence without API confirmation, hard-coded fixtures, or mock reload controls.
- Audit history, conflict UI, or merge resolution; last successful save wins.

## UI scope

- Touches one approved Task board screen.
- Uses existing TaskCard actions for Move left, Move right, Edit, and Delete visibility, but this story implements only move and edit actions.
- Uses EditModal for editing title, description, due date, and status without leaving the one-screen board.
- Uses Field, Button, StatusPill, BoardColumn, StateSummaryCard, Toast, LoadingState, and ErrorState patterns from `design/design-system.md`.
- Move controls must be visible text buttons and keyboard reachable. At first and last status, unavailable move direction must no-op or be disabled to avoid dead controls.
- Edit modal opens with current saved values, focuses title input, supports Cancel and Save, closes on Escape and backdrop click, traps focus while open, and restores focus to opener.
- Validation errors appear inline beside fields, with visible focus indicator and `aria-describedby` for invalid fields.
- Board remains one screen, responsive from 320px upward; columns stack below 900px and render three columns at 900px and above.
- Motion stays subtle: hover/status-change transitions under 200ms, no page-load animation. Add reduced-motion override during implementation for modal/status transitions.

## Acceptance criteria

1. Given existing task title is `Draft`, when Board owner changes title to `Final` and saves, then task card shows `Final` from saved API response.
2. Given existing task has no description, when Board owner adds description `Call supplier` and saves, then task card shows `Call supplier` from saved API response.
3. Given existing task has description `Old`, when Board owner clears description and saves, then task card no longer shows `Old` after API success.
4. Given existing task has no due date, when Board owner adds a due date and saves, then task card shows that due date after API success.
5. Given existing task has a due date, when Board owner clears due date and saves, then task card no longer shows that due date after API success.
6. Given task edits are visible on board, when Board owner reloads the browser, then task card shows exact saved title, description, due date, and status returned by the REST API.
7. Given existing task is in Todo, when Board owner moves it to Doing, then task appears in Doing column after API success.
8. Given existing task is in Doing, when Board owner moves it to Done, then task appears in Done column after API success.
9. Given existing task is in Done, when Board owner moves it to Todo, then task appears in Todo column after API success.
10. Given moved task appears in Done, when Board owner reloads the browser, then task remains in Done from the REST API.
11. Given status counts are visible, when Board owner moves one Todo task to Done, then Todo count decreases by 1 and Done count increases by 1 after API success.
12. Given edited title is blank after trimming whitespace, when Board owner saves, then inline title error appears and no update request is treated as persisted.
13. Given edited title is 120 characters, when Board owner saves, then task is accepted and saved.
14. Given edited title is 121 characters, when Board owner saves, then inline title error names the 120 character limit and task is not saved.
15. Given edited description is 2,000 characters, when Board owner saves, then task is accepted and saved.
16. Given edited description is 2,001 characters, when Board owner saves, then inline description error names the 2,000 character limit and task is not saved.
17. Given edited due date is not a valid calendar date, when Board owner saves, then inline due date error appears and task is not saved.
18. Given edited status is not `todo`, `doing`, or `done`, when Board owner saves or moves, then status error appears, task is not saved, and task stays in previous confirmed column.
19. Given target task no longer exists, when Board owner saves or moves it, then not-found message appears and board refreshes from API without that task.
20. Given two saves affect same task, when both complete, then last successful save wins and board refreshes from saved API response.
21. Given REST API or Postgres update fails, when Board owner saves or moves, then error message appears and task card remains at last confirmed saved values.
22. Given Board owner opens browser storage panels after edit or move, then no task persistence exists in localStorage, sessionStorage, IndexedDB, or cookies.

## Data rules

| Field | Rule |
|---|---|
| Task ID | Required stable identifier for target task update or move. |
| Title | Required; update writes 1 to 120 characters after trimming leading and trailing whitespace. |
| Description | Optional; update writes 0 to 2,000 characters after trimming; blank means absent and is not displayed as old text. |
| Status | Required; update writes exactly one of `todo`, `doing`, `done`. |
| Due date | Optional date-only value; update writes valid calendar date or absent value. |

## Dependencies

- Depends on View tasks by status for board load, three status columns, persisted task cards, empty states, and counts.
- Depends on REST API update endpoint and Postgres persistence defined by architecture/service design.
- Depends on task table with `id`, `title`, optional `description`, `status`, and optional `due_date` fields.
- Depends on approved design and `design/design-system.md` for edit modal, fields, card actions, columns, status pills, counts, loading, error, toast, responsive layout, and accessibility states.
- No external accounts, credentials, or provider setup needed.

## Build notes

- Use REST API as source of truth. Do not use browser storage or fixture data as persisted board state.
- Use full-field update or status-specific update only if service contract provides it; either is acceptable if saved API response drives UI state.
- Save button and move controls should avoid duplicate submissions while request is pending.
- Return external errors useful to UI without exposing database internals.
- Last successful save wins; no conflict-resolution UI required.
