"use client";

import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import type { Task } from "../../lib/api/tasks";
import styles from "./board.module.css";

/**
 * The dialog contract TL's reviews insisted on, in one place: focus moves into
 * the dialog on open, Tab cycles inside it, Escape and the backdrop cancel,
 * and focus returns to the element that opened it. Both the edit modal and the
 * delete confirmation render through this so neither can drift.
 */
function Dialog({ titleId, onClose, children }: { titleId: string; onClose: () => void; children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const opener = document.activeElement as HTMLElement | null;
    const first = ref.current?.querySelector<HTMLElement>("input, textarea, select, button");
    first?.focus();
    return () => opener?.focus();
  }, []);

  function onKeyDown(event: React.KeyboardEvent) {
    if (event.key === "Escape") {
      event.stopPropagation();
      onClose();
      return;
    }
    if (event.key !== "Tab" || !ref.current) return;
    const items = Array.from(
      ref.current.querySelectorAll<HTMLElement>("input, textarea, select, button"),
    ).filter((el) => !el.hasAttribute("disabled"));
    if (items.length === 0) return;
    const [head, tail] = [items[0], items[items.length - 1]];
    if (event.shiftKey && document.activeElement === head) {
      event.preventDefault();
      tail.focus();
    } else if (!event.shiftKey && document.activeElement === tail) {
      event.preventDefault();
      head.focus();
    }
  }

  return (
    <div className={styles.overlay} onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className={styles.dialog} role="dialog" aria-modal="true" aria-labelledby={titleId} ref={ref} onKeyDown={onKeyDown}>
        {children}
      </div>
    </div>
  );
}

export function DeleteTaskDialog({ task, busy, onCancel, onConfirm }: { task: Task; busy: boolean; onCancel: () => void; onConfirm: () => void }) {
  return (
    <Dialog titleId="delete-task-title" onClose={onCancel}>
      <h2 id="delete-task-title">Delete task?</h2>
      <p>
        “{task.title}” will be deleted through the REST API and removed from the database. This cannot be undone.
      </p>
      <div className={styles.dialogActions}>
        <button className={styles.ghost} type="button" onClick={onCancel} disabled={busy}>Cancel</button>
        <button className={styles.dangerSolid} type="button" onClick={onConfirm} disabled={busy}>
          {busy ? "Deleting…" : "Delete task"}
        </button>
      </div>
    </Dialog>
  );
}

export function EditTaskDialog({ task, busy, onCancel, onSave }: {
  task: Task;
  busy: boolean;
  onCancel: () => void;
  onSave: (patch: { title: string; description: string | null; due_date: string | null }) => void;
}) {
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description ?? "");
  const [dueDate, setDueDate] = useState(task.due_date ?? "");
  const titleError = validateTitle(title);

  function submit(event: FormEvent) {
    event.preventDefault();
    if (titleError) return;
    onSave({
      title: title.trim(),
      description: description.trim() === "" ? null : description.trim(),
      due_date: dueDate === "" ? null : dueDate,
    });
  }

  return (
    <Dialog titleId="edit-task-title" onClose={onCancel}>
      <h2 id="edit-task-title">Edit task</h2>
      <form className={styles.field} onSubmit={submit} noValidate>
        <div className={styles.field}>
          <label htmlFor="edit-title">Title</label>
          <input
            id="edit-title"
            className={titleError ? `${styles.input} ${styles.inputInvalid}` : styles.input}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            aria-invalid={titleError ? true : undefined}
            aria-describedby={titleError ? "edit-title-error" : undefined}
            disabled={busy}
          />
          {titleError ? <p className={styles.fieldError} id="edit-title-error">{titleError}</p> : null}
        </div>
        <div className={styles.field}>
          <label htmlFor="edit-description">Description</label>
          <textarea id="edit-description" className={styles.textarea} rows={3} value={description} onChange={(e) => setDescription(e.target.value)} disabled={busy} />
        </div>
        <div className={styles.field}>
          <label htmlFor="edit-due">Due date</label>
          <input id="edit-due" className={styles.input} type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} disabled={busy} />
        </div>
        <div className={styles.dialogActions}>
          <button className={styles.ghost} type="button" onClick={onCancel} disabled={busy}>Cancel</button>
          <button className={styles.primary} type="submit" disabled={busy || titleError !== null}>
            {busy ? "Saving…" : "Save changes"}
          </button>
        </div>
      </form>
    </Dialog>
  );
}

export function validateTitle(raw: string): string | null {
  const title = raw.trim();
  if (title === "") return "Title is required.";
  if ([...title].length > 120) return "Title must be 120 characters or fewer.";
  return null;
}
