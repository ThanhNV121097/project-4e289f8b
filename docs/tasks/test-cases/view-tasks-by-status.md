# Test Cases — View tasks by status

Risk level: medium. Feature reads persisted board state and has one explicit failure path: API load failure. Coverage stays limited to written acceptance criteria plus named failure behavior.

## Case 1
**Scenario**: Render one task in each status column
**Given**: Postgres contains one task with status `todo`, one task with status `doing`, and one task with status `done`
**When**: Board owner opens task board page
**Then**: Three task cards appear and each card is in matching column: `todo` task in Todo, `doing` task in Doing, `done` task in Done

Traceability: TASKS-001 AC-1

## Case 2
**Scenario**: Reload board from API persistence
**Given**: Postgres contains three tasks already shown on board
**When**: Board owner reloads browser
**Then**: Board shows same three tasks from API and no task depends on browser storage for persistence

Traceability: TASKS-001 AC-2

## Case 3
**Scenario**: Empty Doing column shows quiet empty state
**Given**: Postgres contains tasks in Todo and Done, and no tasks in Doing
**When**: Board owner opens task board page
**Then**: Doing column shows quiet empty state and page still renders Todo and Done columns normally

Traceability: TASKS-001 AC-3

## Case 4
**Scenario**: No task persistence in browser storage
**Given**: Postgres contains one task and browser storage panels are open in dev tools
**When**: Board owner inspects localStorage, sessionStorage, IndexedDB, and cookies after page load
**Then**: No task data is present in localStorage, sessionStorage, IndexedDB, or cookies

Traceability: TASKS-001 AC-4

## Case 5
**Scenario**: Task card shows title, description, due date, and status label
**Given**: Postgres contains one task with title `Pay invoice`, description `Due before Friday`, due date `2026-09-01`, and status `doing`
**When**: Board owner opens task board page
**Then**: Task card shows `Pay invoice`, `Due before Friday`, `2026-09-01`, and status label `doing`

Traceability: TASKS-001 AC-5

## Case 6
**Scenario**: Task card hides missing description and due date
**Given**: Postgres contains one task with title `Draft agenda` and status `todo`, with no description and no due date
**When**: Board owner opens task board page
**Then**: Task card shows `Draft agenda` and status label `todo` and does not require description text or due date text

Traceability: TASKS-001 AC-6

## Case 7
**Scenario**: API load failure shows error state instead of stale board data
**Given**: Task list request to REST API fails or times out
**When**: Board owner opens task board page
**Then**: Page shows API error state with retry action and does not show stale browser-storage substitute data

Traceability: TASKS-001 failure behavior: API load failure

## Case 8
**Scenario**: Invalid API status blocks board render
**Given**: REST API returns a task whose status is outside `todo`, `doing`, and `done`
**When**: Board owner opens task board page
**Then**: Board shows API error state and does not render invalid task as a fourth column

Traceability: TASKS-001 failure behavior: invalid API data

## Case 9
**Scenario**: Empty API data still renders all three columns
**Given**: REST API returns an empty task list
**When**: Board owner opens task board page
**Then**: Todo, Doing, and Done columns all render and each column shows quiet empty state

Traceability: TASKS-001 failure behavior: empty data
