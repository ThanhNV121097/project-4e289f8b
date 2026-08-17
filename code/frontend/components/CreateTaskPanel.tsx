"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { createTask, listTasks, Task, TaskStatus } from "../lib/mock/create-task";
import styles from "./CreateTaskPanel.module.css";

const statuses: Array<{ value: TaskStatus; label: string }> = [
  { value: "todo", label: "Todo" },
  { value: "doing", label: "Doing" },
  { value: "done", label: "Done" },
];

type FieldName = "title" | "description" | "dueDate" | "status" | "form";
type Errors = Partial<Record<FieldName, string>>;

export default function CreateTaskPanel() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [saving, setSaving] = useState(false);
  const [toast, setToast] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [status, setStatus] = useState<TaskStatus>("todo");
  const [errors, setErrors] = useState<Errors>({});

  useEffect(() => {
    loadBoard(setTasks, setLoadError, setLoading);
  }, []);

  useEffect(() => {
    if (!toast) return;
    const id = setTimeout(() => setToast(""), 2400);
    return () => clearTimeout(id);
  }, [toast]);

  const counts = useMemo(() => countByStatus(tasks), [tasks]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    const nextErrors = validate(title, description, dueDate, status);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length) return;
    await saveTask({ title, description, dueDate, status, setTasks, setErrors, setSaving, setToast, resetForm });
  }

  function resetForm() {
    setTitle("");
    setDescription("");
    setDueDate("");
    setStatus("todo");
  }

  return (
    <div className={styles.wrap}>
      <header className={styles.hero}>
        <p>Task board</p>
        <h1>One board, three statuses, saved by API</h1>
      </header>
      {loading && <LoadingState />}
      {loadError && <ErrorState onRetry={() => loadBoard(setTasks, setLoadError, setLoading)} />}
      {!loading && !loadError && <Board tasks={tasks} counts={counts} />}
      <form className={styles.form} onSubmit={submit} noValidate>
        <h2>Create task</h2>
        {errors.form && <p className={styles.error}>{errors.form}</p>}
        <Field id="create-title" label="Title" error={errors.title}>
          <input id="create-title" value={title} onChange={(event) => setTitle(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.title)} aria-describedby={errors.title ? "create-title-error" : undefined} />
        </Field>
        <Field id="create-description" label="Description" error={errors.description}>
          <textarea id="create-description" value={description} onChange={(event) => setDescription(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.description)} aria-describedby={errors.description ? "create-description-error" : undefined} />
        </Field>
        <div className={styles.row}>
          <Field id="create-due-date" label="Due date" error={errors.dueDate}>
            <input id="create-due-date" type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.dueDate)} aria-describedby={errors.dueDate ? "create-due-date-error" : undefined} />
          </Field>
          <Field id="create-status" label="Status" error={errors.status}>
            <select id="create-status" value={status} onChange={(event) => setStatus(event.target.value as TaskStatus)} disabled={saving} aria-invalid={Boolean(errors.status)} aria-describedby={errors.status ? "create-status-error" : undefined}>{statuses.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
          </Field>
        </div>
        <button className={styles.primary} disabled={saving}>{saving ? "Saving..." : "Create task"}</button>
      </form>
      {toast && <p className={styles.toast} role="status">{toast}</p>}
    </div>
  );
}

function Board({ tasks, counts }: { tasks: Task[]; counts: Record<TaskStatus, number> }) {
  return <><section className={styles.summary} aria-label="Task totals">{statuses.map(({ value, label }) => <Summary key={value} label={label} count={counts[value]} />)}</section><section className={styles.board} aria-label="Task board">{statuses.map(({ value, label }) => <Column key={value} status={value} label={label} tasks={tasks.filter((task) => task.status === value)} />)}</section></>;
}

function Field({ id, label, error, children }: { id: string; label: string; error?: string; children: React.ReactNode }) {
  return <label htmlFor={id}>{label}{children}{error && <span id={`${id}-error`} className={styles.error}>{error}</span>}</label>;
}

async function loadBoard(setTasks: (tasks: Task[]) => void, setLoadError: (value: boolean) => void, setLoading: (value: boolean) => void) {
  setLoading(true);
  setLoadError(false);
  try {
    setTasks((await listTasks()).tasks);
  } catch {
    setLoadError(true);
  } finally {
    setLoading(false);
  }
}

async function saveTask(args: { title: string; description: string; dueDate: string; status: TaskStatus; setTasks: React.Dispatch<React.SetStateAction<Task[]>>; setErrors: (errors: Errors) => void; setSaving: (value: boolean) => void; setToast: (value: string) => void; resetForm: () => void; }) {
  args.setSaving(true);
  try {
    const saved = await createTask({ title: args.title, description: args.description, due_date: args.dueDate || null, status: args.status });
    args.setTasks((current) => [...current, saved]);
    args.resetForm();
    args.setToast("Task saved.");
  } catch {
    args.setErrors({ form: "Task could not be saved. Try again." });
  } finally {
    args.setSaving(false);
  }
}

function countByStatus(tasks: Task[]) {
  return Object.fromEntries(statuses.map(({ value }) => [value, tasks.filter((task) => task.status === value).length])) as Record<TaskStatus, number>;
}

function validate(title: string, description: string, dueDate: string, status: TaskStatus): Errors {
  const errors: Errors = {};
  if (!title.trim()) errors.title = "Title is required.";
  if (title.trim().length > 120) errors.title = "Title must be 120 characters or fewer.";
  if (description.trim().length > 2000) errors.description = "Description must be 2,000 characters or fewer.";
  if (dueDate && !isDateOnly(dueDate)) errors.dueDate = "Enter a valid due date.";
  if (!statuses.some((item) => item.value === status)) errors.status = "Choose Todo, Doing, or Done.";
  return errors;
}

function isDateOnly(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) return false;
  const date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
  return date.toISOString().slice(0, 10) === value;
}

function LoadingState() {
  return <section className={styles.state} aria-live="polite"><strong>Loading tasks from API...</strong><span /></section>;
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return <section className={styles.state} role="alert"><strong>Cannot load tasks.</strong><p>Retry avoids stale browser data.</p><button className={styles.secondary} onClick={onRetry}>Retry</button></section>;
}

function Summary({ label, count }: { label: string; count: number }) {
  return <article className={styles.card}><p>{label}</p><strong>{count}</strong></article>;
}

function Column({ status, label, tasks }: { status: TaskStatus; label: string; tasks: Task[] }) {
  return <section className={styles.column}><h2><span className={styles[status]} />{label}<b>{tasks.length}</b></h2>{tasks.length ? tasks.map((task) => <TaskCard key={task.id} task={task} />) : <p className={styles.empty}>No tasks yet.</p>}</section>;
}

function TaskCard({ task }: { task: Task }) {
  return <article className={styles.task}><h3>{task.title}</h3>{task.description && <p>{task.description}</p>}{task.due_date && <time dateTime={task.due_date}>{task.due_date}</time>}<span className={`${styles.pill} ${styles[task.status]}`}>{task.status}</span></article>;
}
