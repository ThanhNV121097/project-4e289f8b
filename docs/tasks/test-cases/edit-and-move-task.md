# Test cases — Edit and move task

Risk level: P1. This flow changes saved task data and board placement, so persistence and reload must be proven.

## TASKS-003 — Edit persisted task fields

**Scenario**: Edit title
**Given** Existing task title is `Draft` and task is visible on board
**When** Board owner changes title to `Final` and saves
**Then** Task card shows `Final`

**Scenario**: Add description
**Given** Existing task has no description and task is visible on board
**When** Board owner adds description `Call supplier` and saves
**Then** Task card shows `Call supplier`

**Scenario**: Clear description
**Given** Existing task has description `Old` and task is visible on board
**When** Board owner clears description and saves
**Then** Task card no longer shows `Old`

**Scenario**: Add due date
**Given** Existing task has no due date and task is visible on board
**When** Board owner adds a due date and saves
**Then** Task card shows that due date

**Scenario**: Clear due date
**Given** Existing task has due date and task is visible on board
**When** Board owner clears due date and saves
**Then** Task card no longer shows that due date

**Scenario**: Reload shows saved edit values
**Given** Task edits are saved in Postgres through REST API
**When** Board owner reloads browser
**Then** Task card shows exact saved values from API

**Scenario**: Reject blank edited title
**Given** Existing task is open in edit controls
**When** Board owner saves title made blank after trim
**Then** Inline title error appears and task is not saved

**Scenario**: Accept 120-character edited title
**Given** Existing task is open in edit controls with title field set to 120 characters after trim
**When** Board owner saves
**Then** Task is accepted

**Scenario**: Reject 121-character edited title
**Given** Existing task is open in edit controls with title field set to 121 characters after trim
**When** Board owner saves
**Then** Inline title error names 120 character limit and task is not saved

**Scenario**: Accept 2,000-character edited description
**Given** Existing task is open in edit controls with description field set to 2,000 characters after trim
**When** Board owner saves
**Then** Task is accepted

**Scenario**: Reject 2,001-character edited description
**Given** Existing task is open in edit controls with description field set to 2,001 characters after trim
**When** Board owner saves
**Then** Inline description error names 2,000 character limit and task is not saved

**Scenario**: Reject invalid edited due date
**Given** Existing task is open in edit controls
**When** Board owner saves due date `2026-02-31`
**Then** Inline due date error appears and task is not saved

**Scenario**: Upstream save failure keeps last confirmed values
**Given** Existing task is open in edit controls and API or Postgres update fails
**When** Board owner saves valid edits
**Then** Error message appears and task card remains at last confirmed saved values

**Scenario**: Not found on edit
**Given** Target task no longer exists in API data
**When** Board owner saves edits for that task
**Then** Not-found message appears and board refreshes from API

**Scenario**: Concurrent edit last save wins
**Given** Two saves affect same task and later save succeeds last
**When** Board owner refreshes board from API
**Then** Board shows values from last successful save

## TASKS-004 — Move persisted task between statuses

**Scenario**: Move todo to doing
**Given** Existing task is in Todo column
**When** Board owner moves it to Doing
**Then** Task appears in Doing column

**Scenario**: Move doing to done
**Given** Existing task is in Doing column
**When** Board owner moves it to Done
**Then** Task appears in Done column

**Scenario**: Move done to todo
**Given** Existing task is in Done column
**When** Board owner moves it to Todo
**Then** Task appears in Todo column

**Scenario**: Reload keeps moved status
**Given** Moved task appears in Done and saved status is in Postgres through REST API
**When** Board owner reloads browser
**Then** Task remains in Done from API

**Scenario**: Status counts update on move
**Given** Status counts are visible and one Todo task is on board
**When** Board owner moves that task to Done
**Then** Todo count decreases by 1 and Done count increases by 1

**Scenario**: Reject invalid edited status
**Given** Existing task is open in edit controls
**When** Board owner saves status not in `todo`, `doing`, or `done`
**Then** Status error appears, task is not saved, and task stays in previous column

**Scenario**: Upstream move failure keeps last confirmed column
**Given** Existing task is visible in its last confirmed column and API or Postgres update fails
**When** Board owner moves task to another status
**Then** Error message appears and task card remains at last confirmed saved values

**Scenario**: Not found on move
**Given** Target task no longer exists in API data
**When** Board owner moves that task
**Then** Not-found message appears and board refreshes from API

**Scenario**: Concurrent move last save wins
**Given** Two saves affect same task and later save succeeds last
**When** Board owner refreshes board from API
**Then** Board shows task in status from last successful save
