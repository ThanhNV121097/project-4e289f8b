# Story — Delete task

Module: `tasks`
Plan item: Delete task
Requirement: TASKS-005 — Delete persisted task

## User story

As a Board owner, I want to delete a task, so that unwanted work is removed from Postgres and stays removed after browser reload.

## In scope

- Delete action on each persisted task card.
- Confirmation before delete is sent.
- Cancel path that leaves task visible and sends no API request.
- Confirm path that deletes the task through the REST API.
- Board update after API success using confirmed saved state, not browser storage.
- Affected column count update after deletion.
- Quiet empty state when deletion leaves a status column empty.
- Browser reload after deletion must not show the deleted task from the API.
- Not-found handling when the target task no longer exists.
- Upstream delete failure handling that keeps last confirmed saved task visible.

## Out of scope

- Sign-up, sign-in, sessions, accounts, users, roles, permissions, or ownership checks beyond the single Board owner actor.
- Multiple boards, projects, members, assignments, comments, labels, tags, activity log, file upload, search, notifications, archive, or undo.
- Bulk delete, drag-to-delete, soft delete, trash, restore, or retention policy.
- Deleting through browser storage, client fixtures, cookies, URL parameters, or any non-Postgres persistence path.
- Changing title, description, due date, or status; those belong to create/edit/move stories.
- Mockup-only “Simulate reload from API” button and blue “Persistence rule” banner.

## UI scope

Touches one approved Task board screen.

- TaskCard: includes visible `Delete` danger mini button for each task.
- Delete confirmation: shown before deletion, displays task title, and offers confirm and cancel controls.
- BoardColumn: removes card only after REST API success and shows `No {status} tasks.` when column becomes empty.
- StateSummaryCard: updates affected status count after confirmed deletion.
- Error feedback: shows not-found or delete-failure message without exposing database internals.
- Accessibility: delete, confirm, and cancel controls are keyboard reachable with visible focus. Confirmation supports Escape/backdrop cancel if implemented as modal, matching EditModal behavior.

## Acceptance criteria

1. Given an existing task is visible, when Board owner chooses `Delete`, then confirmation is shown before any delete request is sent.
2. Given delete confirmation is shown, when Board owner cancels, then task remains visible and no REST API delete request is sent.
3. Given delete confirmation is shown, when Board owner confirms, then system sends delete request for that task ID through the REST API.
4. Given the REST API confirms deletion, when board updates, then deleted task disappears from its status column.
5. Given deleted task was the only task in `done`, when REST API confirms deletion, then Done column shows `No done tasks.` quiet empty state and Done count is `0`.
6. Given task disappears after confirmed delete, when Board owner reloads browser, then deleted task does not appear in data loaded from the API.
7. Given target task no longer exists, when Board owner confirms delete, then board refreshes from the API, removes the stale task, and shows a not-found message.
8. Given REST API or Postgres delete fails, when Board owner confirms delete, then error message appears and task remains visible as last confirmed saved data.
9. Given Board owner attempts delete on a task already removed from current data, when delete is submitted, then not-found message appears and board refreshes from the API.
10. Given browser dev tools storage panels are inspected after delete, then deleted task is not persisted in localStorage, sessionStorage, IndexedDB, cookies, or URL parameters.

## Dependencies

- View tasks by status story supplies loaded task cards, status columns, counts, loading/error handling, and quiet empty states.
- REST API must support deleting a task by stable task ID.
- Postgres must be source of task truth.
- Design system tokens and components in `design/design-system.md` must be followed.
- Architecture overview requires Next.js UI, Go REST API, PostgreSQL persistence, parameterized SQL, request validation, and no browser-storage persistence.

## No blocking questions

No stakeholder decision blocks this story. Confirmation is required by SRS. Delete is permanent; undo/restore is out of scope.
