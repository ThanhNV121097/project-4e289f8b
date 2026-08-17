# Test cases — Create task

Risk: high. This story writes persisted data and must prove API save + reload from Postgres.

## Coverage check
Acceptance criteria covered: AC-1..AC-6.
Named failure / boundary cases covered: blank title, title 120, title 121, description 2000, description 2001, invalid due date, invalid status, save failure, duplicate title allowed, board owner allowed.

---

**Scenario**: Create task with default status
**Given** Create form is empty
**When** Board owner enters title `Pay invoice` and submits
**Then** Task `Pay invoice` appears in Todo column

**Scenario**: Reload shows created task from API
**Given** Task `Pay invoice` was created and appears on board
**When** Board owner reloads browser
**Then** Same task appears from API after reload

**Scenario**: Create task with description
**Given** Create form is empty and description field contains `Due before Friday`
**When** Board owner enters title `Pay invoice` and submits
**Then** Created task card shows `Due before Friday`

**Scenario**: Create task with due date
**Given** Create form is empty and due date field contains `2026-08-17`
**When** Board owner enters title `Pay invoice` and submits
**Then** Created task card shows `2026-08-17`

**Scenario**: Create task with initial Doing status
**Given** Create form is empty and initial status is set to `doing`
**When** Board owner enters title `Pay invoice` and submits
**Then** Created task appears in Doing column

**Scenario**: Create task without description or due date
**Given** Create form is empty
**When** Board owner enters title `Pay invoice` and submits with description blank and due date blank
**Then** Task is saved without description and due date

**Scenario**: Reject blank title
**Given** Create form is empty
**When** Board owner enters only whitespace in title and submits
**Then** Inline title error appears and task is not saved

**Scenario**: Accept title at 120 characters
**Given** Create form is empty and title field has exactly 120 trimmed characters
**When** Board owner submits create form
**Then** Task is accepted and saved

**Scenario**: Reject title at 121 characters
**Given** Create form is empty and title field has 121 trimmed characters
**When** Board owner submits create form
**Then** Inline title error names 120 character limit and task is not saved

**Scenario**: Accept description at 2000 characters
**Given** Create form is empty and description field has exactly 2000 trimmed characters
**When** Board owner enters title `Pay invoice` and submits
**Then** Task is accepted and saved

**Scenario**: Reject description at 2001 characters
**Given** Create form is empty and description field has 2001 trimmed characters
**When** Board owner enters title `Pay invoice` and submits
**Then** Inline description error names 2000 character limit and task is not saved

**Scenario**: Reject invalid due date
**Given** Create form is empty and due date field contains `2026-02-30`
**When** Board owner enters title `Pay invoice` and submits
**Then** Inline due date error appears and task is not saved

**Scenario**: Reject invalid initial status
**Given** Create form is empty and initial status value is invalid
**When** Board owner enters title `Pay invoice` and submits
**Then** Status error appears and task is not saved

**Scenario**: Save failure keeps task off board
**Given** Create form is empty and API save fails
**When** Board owner enters title `Pay invoice` and submits
**Then** Error message appears and task is not added to board as persisted data

**Scenario**: Duplicate title is allowed
**Given** Another task already exists with title `Pay invoice`
**When** Board owner enters title `Pay invoice` and submits
**Then** Task is saved and duplicate title is allowed
