"use client";

import { useState, type FormEvent, type ReactNode } from "react";
import type { TaskInput, TaskStatus } from "../../lib/api/tasks";
import { validateTitle } from "./Dialogs";
import styles from "./board.module.css";

function validateDescription(raw: string): string | null {
  return [...raw.trim()].length > 2000 ? "Description must be 2,000 characters or fewer." : null;
}

/** A labelled control with its error linked the way TL's review demanded:
 * aria-invalid on the control, aria-describedby naming the error element. */
function Field({ id, label, error, children }: { id: string; label: string; error: string | null; children: (aria: { "aria-invalid"?: true; "aria-describedby"?: string; className: string }) => ReactNode }) {
  const invalidClass = error ? ` ${styles.inputInvalid}` : "";
  return (
    <div className={styles.field}>
      <label htmlFor={id}>{label}</label>
      {children({
        "aria-invalid": error ? true : undefined,
        "aria-describedby": error ? `${id}-error` : undefined,
        className: invalidClass,
      })}
      {error ? <p className={styles.fieldError} id={`${id}-error`}>{error}</p> : null}
    </div>
  );
}

/**
 * The Create-task story's form, integrated onto the one board. A failed
 * submit keeps every value the user typed.
 */
export default function CreateTaskForm({ busy, onCreate }: { busy: boolean; onCreate: (input: TaskInput) => Promise<boolean> }) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [status, setStatus] = useState<TaskStatus>("todo");
  const [showErrors, setShowErrors] = useState(false);

  const titleError = showErrors ? validateTitle(title) : null;
  const descriptionError = showErrors ? validateDescription(description) : null;

  async function submit(event: FormEvent) {
    event.preventDefault();
    setShowErrors(true);
    if (validateTitle(title) || validateDescription(description)) return;
    const created = await onCreate({
      title: title.trim(),
      description: description.trim() === "" ? null : description.trim(),
      status,
      due_date: dueDate === "" ? null : dueDate,
    });
    if (!created) return;
    setTitle("");
    setDescription("");
    setDueDate("");
    setStatus("todo");
    setShowErrors(false);
  }

  return (
    <section className={styles.formPanel} aria-labelledby="create-task-title">
      <h2 id="create-task-title">Create task</h2>
      <p className={styles.formHint}>Title is required. Description and due date are optional. Saved through the REST API.</p>
      <form className={styles.field} onSubmit={submit} noValidate>
        <div className={styles.formGrid}>
          <Field id="create-title" label="Title" error={titleError}>
            {(aria) => (
              <input id="create-title" {...aria} className={styles.input + aria.className} value={title} onChange={(e) => setTitle(e.target.value)} disabled={busy} />
            )}
          </Field>
          <Field id="create-description" label="Description" error={descriptionError}>
            {(aria) => (
              <textarea id="create-description" {...aria} className={styles.textarea + aria.className} rows={2} value={description} onChange={(e) => setDescription(e.target.value)} disabled={busy} />
            )}
          </Field>
          <Field id="create-due" label="Due date" error={null}>
            {(aria) => (
              <input id="create-due" {...aria} className={styles.input + aria.className} type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} disabled={busy} />
            )}
          </Field>
          <Field id="create-status" label="Status" error={null}>
            {(aria) => (
              <select id="create-status" {...aria} className={styles.select + aria.className} value={status} onChange={(e) => setStatus(e.target.value as TaskStatus)} disabled={busy}>
                <option value="todo">Todo</option>
                <option value="doing">Doing</option>
                <option value="done">Done</option>
              </select>
            )}
          </Field>
        </div>
        <div className={styles.formActions}>
          <button className={styles.primary} type="submit" disabled={busy}>
            {busy ? "Adding…" : "Add task"}
          </button>
        </div>
      </form>
    </section>
  );
}
