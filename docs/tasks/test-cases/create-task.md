# Test Cases — Create task

Risk level: high. This flow writes persisted board state and must survive API save + browser reload.

## Automated coverage

### Scenario: create task with title shows on Todo board
**Given** create form is empty and Postgres has no task titled `Pay invoice`
**When** Board owner enters title `Pay invoice` and submits
**Then** task `Pay invoice` appears in Todo column and saved state comes from API response

### Scenario: created task survives browser reload
**Given** task `Pay invoice` was created and board shows it in Todo
**When** Board owner reloads browser
**Then** same task appears from API in Todo column

### Scenario: create task keeps description
**Given** create form is empty
**When** Board owner enters title `Pay invoice`, description `Due before Friday`, and submits
**Then** created task card shows `Due before Friday`

### Scenario: create task keeps due date
**Given** create form is empty
**When** Board owner enters title `Pay invoice`, chooses due date `2026-09-30`, and submits
**Then** created task card shows due date `2026-09-30`

### Scenario: create task honors initial Doing status
**Given** create form is empty
**When** Board owner enters title `Pay invoice`, chooses initial status `doing`, and submits
**Then** created task appears in Doing column

### Scenario: create task stores blank optional fields as absent
**Given** create form is empty
**When** Board owner enters title `Pay invoice` and leaves description and due date blank, then submits
**Then** task is saved without description and without due date

### Scenario: blank title blocks save
**Given** create form is empty
**When** Board owner submits title that is blank after trimming whitespace
**Then** inline title error appears and task is not saved

### Scenario: title at 120 chars is accepted
**Given** create form is empty
**When** Board owner submits title with exactly 120 characters
**Then** task is accepted and saved

### Scenario: title at 121 chars is rejected
**Given** create form is empty
**When** Board owner submits title with 121 characters
**Then** inline title error names 120 character limit and task is not saved

### Scenario: description at 2,000 chars is accepted
**Given** create form is empty
**When** Board owner submits title `Pay invoice` with description of exactly 2,000 characters
**Then** task is accepted and saved

### Scenario: description at 2,001 chars is rejected
**Given** create form is empty
**When** Board owner submits title `Pay invoice` with description of 2,001 characters
**Then** inline description error names 2,000 character limit and task is not saved

### Scenario: invalid due date blocks save
**Given** create form is empty
**When** Board owner enters title `Pay invoice` and invalid due date `2026-02-30`, then submits
**Then** inline due date error appears and task is not saved

### Scenario: invalid initial status blocks save
**Given** create form is empty
**When** Board owner enters title `Pay invoice` and an initial status outside `todo`, `doing`, `done`, then submits
**Then** status error appears and task is not saved

### Scenario: API or Postgres save failure leaves task unsaved
**Given** create form has valid title `Pay invoice`
**When** REST API save fails or Postgres is unavailable during submit
**Then** error message appears and task is not added to board as persisted data

### Scenario: duplicate title is allowed
**Given** another task already has title `Pay invoice`
**When** Board owner creates another task with title `Pay invoice`
**Then** task is saved successfully and both tasks remain allowed

## Manual coverage

### Scenario: created task persists only in API-backed board state
**Given** task `Pay invoice` was created
**When** Board owner inspects browser storage panels
**Then** no task persistence is present in localStorage, sessionStorage, IndexedDB, or cookies
