# Test Cases — Delete task

Risk level: medium. One delete flow, but it changes persisted data and has confirm/cancel, reload, empty-state, and failure behavior.

## Coverage check
- TASKS-005 AC-1: confirmation shown before deletion
- TASKS-005 AC-2: cancel keeps task visible
- TASKS-005 AC-3: confirm removes task from board
- TASKS-005 AC-4: reload keeps deleted task gone from API
- TASKS-005 AC-5: deleting last Done task shows quiet empty state
- Failure cases: not found, cancelled action, upstream failure, repeated delete
- Permission: Board owner allowed; no other actor in scope

## Scenario: Delete shows confirmation
**Given** one existing task is visible on board
**When** Board owner chooses delete on task
**Then** confirmation dialog appears before any deletion request is sent

Requirement traceability: TASKS-005 AC-1

## Scenario: Cancel delete keeps task visible
**Given** delete confirmation dialog is shown for existing task
**When** Board owner cancels confirmation
**Then** task remains visible on board and no delete request is sent

Requirement traceability: TASKS-005 AC-2; failure case: cancelled action

## Scenario: Confirm delete removes task from board
**Given** existing task is visible and delete confirmation dialog is shown
**When** Board owner confirms deletion
**Then** task is deleted through REST API and disappears from board after API success

Requirement traceability: TASKS-005 AC-3

## Scenario: Reload keeps deleted task gone
**Given** task was deleted and disappeared from board
**When** Board owner reloads browser
**Then** deleted task does not appear from API

Requirement traceability: TASKS-005 AC-4; persistence requirement

## Scenario: Delete last Done task leaves Done empty state
**Given** only task in Done column is visible and delete confirmation dialog is shown
**When** Board owner confirms deletion
**Then** Done column shows quiet empty state

Requirement traceability: TASKS-005 AC-5

## Scenario: Delete target already missing shows not-found
**Given** target task no longer exists on server
**When** Board owner tries to delete task
**Then** board shows not-found message and refreshes from API

Requirement traceability: TASKS-005 failure case Not found

## Scenario: Delete cancel sends no API request
**Given** delete confirmation dialog is shown
**When** Board owner cancels confirmation
**Then** no API delete request is sent and task remains visible

Requirement traceability: TASKS-005 failure case Cancelled action

## Scenario: Delete API failure keeps task visible
**Given** existing task is visible and delete confirmation dialog is shown
**When** Board owner confirms deletion but REST API or Postgres delete fails
**Then** error message appears and task remains visible as last confirmed saved data

Requirement traceability: TASKS-005 failure case Upstream failure

## Scenario: Repeated delete on removed task shows not-found
**Given** task was already removed from current data
**When** Board owner attempts delete again
**Then** board shows not-found message and refreshes from API

Requirement traceability: TASKS-005 failure case Repeated delete

## Scenario: Board owner can delete task
**Given** existing task is visible
**When** Board owner chooses delete
**Then** delete action is available; no denied-role case exists in this module

Requirement traceability: TASKS-005 permission behavior
