# SRS — Tasks

Module: `tasks`
Last updated: 2026-08-17
Design: [View Design](http://localhost:8080/design/4e289f8b-2119-4692-af5e-3d34d4677f09)
Design system: `design/design-system.md`

## 1. Purpose

The tasks module lets one person manage one task board on one screen. It is the whole product surface for "Task board for one person": viewing, creating, editing, moving, and deleting tasks backed by Postgres through a REST API.

If this module does not exist, the stakeholder cannot prove persistence: adding a task, reloading the browser, moving it to Done, and reloading again must show database state exactly.

## 2. Actors

| Actor | Who they are | What they may do in this module |
|---|---|---|
| Board owner | The one person using the task board, with no sign-in or account flow | View all tasks, create tasks, edit task title, edit optional description, edit due date, move tasks among todo / doing / done, delete tasks |

Permission rule: there is no authentication, account ownership, multi-user access, member role, admin role, assignment, or board selection in this module. Every available task action is available to the Board owner.

## 3. Scope

**In scope** — the functions specified below, by their plan titles:

- View tasks by status
- Create task
- Edit and move task
- Delete task

**Out of scope** — deliberately not built for this project:

- Sign-up, sign-in, sessions, accounts, users, roles, and permissions beyond the single Board owner actor.
- Multiple users, members, assignments, ownership transfer, and sharing.
- Comments, labels, tags, activity log, file upload, search, notifications, multiple boards, multiple projects, and archived boards.
- Browser storage as task persistence. Reloaded data must come from Postgres behind the REST API.
- Mockup-only "Simulate reload from API" button and blue "Persistence rule" banner. Browser reload is the persistence test.

## 4. Functional requirements

### 4.1 View tasks by status

**Requirement TASKS-001 — Load persisted tasks into three status columns**

*As a* Board owner, *I want to* see every persisted task grouped by status, *so that* I know what is todo, doing, and done after any browser reload.

Behaviour:

1. The Board owner opens the task board page.
2. The system fetches tasks from the REST API before rendering task cards as current board state.
3. The system renders exactly three columns: Todo, Doing, and Done.
4. The system places each task in the column matching its saved status: `todo`, `doing`, or `done`.
5. The system shows task title, optional description text when present, due date when present, current status label, and task actions on each card.
6. The system shows a quiet empty state inside any status column with zero tasks.
7. The system does not read tasks from localStorage, sessionStorage, IndexedDB, cookies, URL parameters, or hard-coded client fixtures as persisted board state.

**Acceptance criteria** — each maps one-to-one onto a test case in `docs/tasks/test-cases/view-tasks-by-status.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Postgres contains one `todo`, one `doing`, and one `done` task | Board owner opens the page | Each task appears in its matching status column |
| AC-2 | Postgres contains three tasks | Board owner reloads the browser | The page shows those same three tasks from the API |
| AC-3 | Postgres contains no `doing` tasks | Board owner opens the page | Doing column shows a quiet empty state |
| AC-4 | Postgres contains one task | Board owner opens browser dev tools storage panels | No task persistence is present in localStorage, sessionStorage, IndexedDB, or cookies |
| AC-5 | Postgres contains a task with description and due date | Board owner opens the page | Task card shows title, description, due date, and status label |
| AC-6 | Postgres contains a task without description and without due date | Board owner opens the page | Task card shows title and status without requiring description or due date text |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| API load failure | Task list request fails or times out | Board shows API error state with retry action; no stale browser-storage substitute is shown |
| Invalid API data | A returned task has status outside `todo`, `doing`, `done` | Task is not rendered as a fourth column; board shows API error state because persisted data violates allowed statuses |
| Empty data | API returns an empty list | All three columns render and each column shows quiet empty state |
| Not permitted | Actor is Board owner | Action is allowed; no other actor exists in scope |
| Upstream unavailable | REST API or Postgres is unavailable | Board remains usable enough to retry loading and does not claim data was loaded |

**Data touched**

| Field | Type | Required | Rule |
|---|---|---|---|
| Task ID | identifier | yes | Stable identifier used for edit, move, and delete actions; visible UI need not display it |
| Title | text | yes | 1 to 120 characters after trimming leading and trailing whitespace |
| Description | text | no | 0 to 2,000 characters after trimming leading and trailing whitespace; blank is stored and displayed as absent |
| Status | enum | yes | Exactly one of `todo`, `doing`, `done` |
| Due date | date | no | Calendar date with no time-of-day requirement; blank is stored and displayed as absent |

### 4.2 Create task

**Requirement TASKS-002 — Create persisted task**

*As a* Board owner, *I want to* create a task with title, optional description, due date, and initial status, *so that* new work is saved in Postgres and survives browser reload.

Behaviour:

1. The Board owner enters a title.
2. The Board owner may enter a description.
3. The Board owner may choose a due date.
4. The Board owner chooses an initial status; default initial status is `todo` when no status is changed by the Board owner.
5. The Board owner submits the create form.
6. The system validates fields before save.
7. The system saves the task through the REST API.
8. The system updates the board from the saved result returned by the API.
9. Browser reload shows the created task from Postgres in the saved status column.

**Acceptance criteria** — each maps one-to-one onto a test case in `docs/tasks/test-cases/create-task.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Create form is empty | Board owner enters title `Pay invoice` and submits | Task `Pay invoice` appears in Todo column |
| AC-2 | Created task appears on board | Board owner reloads the browser | Same task appears from the API after reload |
| AC-3 | Board owner entered description `Due before Friday` | Board owner creates the task | Created task card shows `Due before Friday` |
| AC-4 | Board owner chooses a due date | Board owner creates the task | Created task card shows that due date |
| AC-5 | Board owner chooses initial status `doing` | Board owner creates the task | Created task appears in Doing column |
| AC-6 | Board owner leaves description and due date blank | Board owner creates a titled task | Task is saved without description and due date |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Title is blank after trimming whitespace | Inline title error appears; task is not saved |
| Boundary | Title is 120 characters | Task is accepted |
| Boundary | Title is 121 characters | Inline title error names the 120 character limit; task is not saved |
| Boundary | Description is 2,000 characters | Task is accepted |
| Boundary | Description is 2,001 characters | Inline description error names the 2,000 character limit; task is not saved |
| Invalid input | Due date is not a valid calendar date | Inline due date error appears; task is not saved |
| Invalid input | Initial status is not `todo`, `doing`, or `done` | Status error appears; task is not saved |
| Not permitted | Actor is Board owner | Action is allowed; no other actor exists in scope |
| Upstream failure | REST API or Postgres save fails | Error message appears; task is not added to board as persisted data |
| Duplicate title | Another task has same title | Task is allowed; titles are not unique |

**Data touched**

| Field | Type | Required | Rule |
|---|---|---|---|
| Title | text | yes | Create writes 1 to 120 trimmed characters |
| Description | text | no | Create writes 0 to 2,000 trimmed characters; blank means absent |
| Status | enum | yes | Create writes one of `todo`, `doing`, `done`; default is `todo` |
| Due date | date | no | Create writes a valid calendar date or absent value |

### 4.3 Edit and move task

**Requirement TASKS-003 — Edit persisted task fields**

*As a* Board owner, *I want to* edit a task title, optional description, and due date, *so that* saved task details stay accurate after browser reload.

Behaviour:

1. The Board owner opens edit controls for an existing task.
2. The system shows current saved title, description, due date, and status in editable controls.
3. The Board owner changes title, description, due date, or status.
4. The Board owner saves changes.
5. The system validates fields before save.
6. The system saves changes through the REST API.
7. The system updates the board from the saved result returned by the API.
8. Browser reload shows exact saved values from Postgres.

**Acceptance criteria** — each maps one-to-one onto a test case in `docs/tasks/test-cases/edit-and-move-task.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Existing task title is `Draft` | Board owner changes title to `Final` and saves | Task card shows `Final` |
| AC-2 | Existing task has no description | Board owner adds description `Call supplier` and saves | Task card shows `Call supplier` |
| AC-3 | Existing task has description `Old` | Board owner clears description and saves | Task card no longer shows `Old` |
| AC-4 | Existing task has no due date | Board owner adds a due date and saves | Task card shows that due date |
| AC-5 | Existing task has due date | Board owner clears due date and saves | Task card no longer shows that due date |
| AC-6 | Task edits are visible on board | Board owner reloads the browser | Task card shows exact saved values from the API |

**Requirement TASKS-004 — Move persisted task between statuses**

*As a* Board owner, *I want to* move a task between Todo, Doing, and Done, *so that* workflow status is saved in Postgres and survives browser reload.

Behaviour:

1. The Board owner chooses a status change from a task card or edit controls.
2. The system saves the new status through the REST API.
3. The system moves the task card to the column matching the saved status returned by the API.
4. The system updates status counts for Todo, Doing, and Done.
5. Browser reload shows the task in the saved status column from Postgres.

**Acceptance criteria** — each maps one-to-one onto a test case in `docs/tasks/test-cases/edit-and-move-task.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-7 | Existing task is in Todo | Board owner moves it to Doing | Task appears in Doing column |
| AC-8 | Existing task is in Doing | Board owner moves it to Done | Task appears in Done column |
| AC-9 | Existing task is in Done | Board owner moves it to Todo | Task appears in Todo column |
| AC-10 | Moved task appears in Done | Board owner reloads the browser | Task remains in Done from the API |
| AC-11 | Status counts are visible | Board owner moves one Todo task to Done | Todo count decreases by 1 and Done count increases by 1 |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Edited title is blank after trimming whitespace | Inline title error appears; task is not saved |
| Boundary | Edited title is 120 characters | Task is accepted |
| Boundary | Edited title is 121 characters | Inline title error names the 120 character limit; task is not saved |
| Boundary | Edited description is 2,000 characters | Task is accepted |
| Boundary | Edited description is 2,001 characters | Inline description error names the 2,000 character limit; task is not saved |
| Invalid input | Edited due date is not a valid calendar date | Inline due date error appears; task is not saved |
| Invalid input | Edited status is not `todo`, `doing`, or `done` | Status error appears; task is not saved and task stays in previous column |
| Not found | Target task no longer exists | Not-found message appears; task is removed from current board after refresh from API |
| Conflict | Two saves affect the same task | Last successful save wins; board refreshes from saved API response |
| Not permitted | Actor is Board owner | Action is allowed; no other actor exists in scope |
| Upstream failure | REST API or Postgres update fails | Error message appears; task card remains at last confirmed saved values |

**Data touched**

| Field | Type | Required | Rule |
|---|---|---|---|
| Task ID | identifier | yes | Identifies target task to edit or move |
| Title | text | yes | Update writes 1 to 120 trimmed characters |
| Description | text | no | Update writes 0 to 2,000 trimmed characters; blank means absent |
| Status | enum | yes | Update writes exactly one of `todo`, `doing`, `done` |
| Due date | date | no | Update writes a valid calendar date or absent value |

### 4.4 Delete task

**Requirement TASKS-005 — Delete persisted task**

*As a* Board owner, *I want to* delete a task, *so that* unwanted work is removed from Postgres and stays removed after browser reload.

Behaviour:

1. The Board owner chooses delete on an existing task.
2. The system asks for confirmation before deletion.
3. The Board owner confirms deletion.
4. The system deletes the task through the REST API.
5. The system removes the task from the board after API success.
6. Browser reload does not show the deleted task.

**Acceptance criteria** — each maps one-to-one onto a test case in `docs/tasks/test-cases/delete-task.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Existing task is visible | Board owner chooses delete | Confirmation is shown before deletion |
| AC-2 | Delete confirmation is shown | Board owner cancels | Task remains visible |
| AC-3 | Delete confirmation is shown | Board owner confirms | Task disappears from the board |
| AC-4 | Task has disappeared after delete | Board owner reloads the browser | Deleted task does not appear from the API |
| AC-5 | Deleted task was only task in Done | Board owner confirms delete | Done column shows quiet empty state |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not found | Target task no longer exists | Board removes the task after refresh from API and shows not-found message |
| Cancelled action | Board owner cancels confirmation | No API delete is sent; task remains visible |
| Not permitted | Actor is Board owner | Action is allowed; no other actor exists in scope |
| Upstream failure | REST API or Postgres delete fails | Error message appears; task remains visible as last confirmed saved data |
| Repeated delete | Board owner attempts to delete a task already removed from current data | Not-found message appears; board refreshes from API |

**Data touched**

| Field | Type | Required | Rule |
|---|---|---|---|
| Task ID | identifier | yes | Identifies target task to delete |
| Title | text | yes | Read for confirmation display |
| Status | enum | yes | Read to update affected column count after delete |

## 5. Screens

Design section: [View Design](http://localhost:8080/design/4e289f8b-2119-4692-af5e-3d34d4677f09)

Approved design palette: `#2563EB` primary, `#F8FAFC` background, `#0F172A` text, `#D97706` doing, `#059669` done. UI uses neutral grey/slate surfaces, blue primary action, muted status tints, no gradients, no heavy color blocks, subtle hover/status-change transitions under 200ms, no page-load animation.

| Screen | Section in the design | Functions it serves | States that must exist |
|---|---|---|---|
| Task board | One responsive screen with topbar, hero, task totals, create form, API loading/error states, three board columns, task cards, edit modal, move actions, delete action, quiet empty states | TASKS-001, TASKS-002, TASKS-003, TASKS-004, TASKS-005 | default, API loading, API error with retry, all columns empty, one column empty, create validation error, edit validation error, delete confirmation, save failure |

## 6. Non-functional requirements

| Area | Requirement |
|---|---|
| Persistence | Browser reload after create, edit, move, or delete must render state returned by the REST API backed by Postgres; browser storage must not be source of task truth |
| Performance | Initial task list renders within 2 seconds for 200 tasks on local development environment after API response starts |
| Accessibility | Keyboard can reach create, edit, move, delete, confirmation, cancel, save, and retry controls; focus indicator remains visible; form inputs have labels; text contrast is at least 4.5:1 |
| Responsive | One screen works from 320px width upward with no horizontal page scroll; below 900px status columns stack vertically; at 900px and above columns render in three-column board layout |
| Motion | Hover and status-change transitions complete within 200ms; no page-load animation is used |
| Localisation | Product copy is English; due dates display as calendar dates without time |
| Privacy | Stored data is limited to task title, optional description, status, due date, and generated task identifier |

## 7. Dependencies and assumptions

- **Depends on:** REST API, for reading and writing tasks.
- **Depends on:** Postgres, for durable task persistence.
- **Depends on:** Approved design and `design/design-system.md`, for UI layout, colors, states, and components.
- **Assumption:** Single Board owner has local access to the board without authentication. If authentication is later required, it is new scope and must not be added silently.
- **Assumption:** Due date is date-only. If time-of-day or timezone-aware deadlines are later required, data model and UI must change.
- **Assumption:** Last successful save wins for concurrent edits because multiple users and activity history are explicitly out of scope.

| Open question | Proposed default | Who decides |
|---|---|---|
| None | No open product questions for this module | Stakeholder |

## 8. Traceability

| Plan item | Requirement ids | Test cases |
|---|---|---|
| View tasks by status | TASKS-001 | `docs/tasks/test-cases/view-tasks-by-status.md` |
| Create task | TASKS-002 | `docs/tasks/test-cases/create-task.md` |
| Edit and move task | TASKS-003, TASKS-004 | `docs/tasks/test-cases/edit-and-move-task.md` |
| Delete task | TASKS-005 | `docs/tasks/test-cases/delete-task.md` |
