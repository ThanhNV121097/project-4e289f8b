"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { createTask, listTasks, Task, TaskStatus } from "../lib/mock/create-task";
import styles from "./CreateTaskPanel.module.css";

const statuses: Array<{ value: TaskStatus; label: string }> = [
  { value: "todo", label: "Todo" },
  { value: "doing", label: "Doing" },
  { value: "done", label: "Done" },
];

type Errors = Partial<Record<"title" | "description" | "dueDate" | "status" | "form", string>>;

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
    listTasks()
      .then((data) => setTasks(data.tasks))
      .catch(() => setLoadError(true))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!toast) return;
    const id = setTimeout(() => setToast(""), 2400);
    return () => clearTimeout(id);
  }, [toast]);

  const counts = useMemo(
    () => Object.fromEntries(statuses.map(({ value }) => [value, tasks.filter((task) => task.status === value).length])),
    [tasks],
  ) as Record<TaskStatus, number>;

  async function submit(event: FormEvent) {
    event.preventDefault();
    const nextErrors = validate(title, description, dueDate, status);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length) return;
    setSaving(true);
    try {
      const saved = await createTask({ title, description, due_date: dueDate || null, status });
      setTasks((current) => [...current, saved]);
      setTitle("");
      setDescription("");
      setDueDate("");
      setStatus("todo");
      setToast("Task saved.");
    } catch {
      setErrors({ form: "Task could not be saved. Try again." });
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className={styles.wrap}>
      <header className={styles.hero}>
        <p>Task board</p>
        <h1>One board, three statuses, saved by API</h1>
      </header>
      {loading && <LoadingState />}
      {loadError && <ErrorState />}
      {!loading && !loadError && (
        <>
          <section className={styles.summary} aria-label="Task totals">
            {statuses.map(({ value, label }) => <Summary key={value} label={label} count={counts[value]} />)}
          </section>
          <section className={styles.board} aria-label="Task board">
            {statuses.map(({ value, label }) => (
              <Column key={value} status={value} label={label} tasks={tasks.filter((task) => task.status === value)} />
            ))}
          </section>
        </>
      )}
      <form className={styles.form} onSubmit={submit} noValidate>
        <h2>Create task</h2>
        {errors.form && <p className={styles.error}>{errors.form}</p>}
        <label>Title<input value={title} onChange={(event) => setTitle(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.title)} /></label>
        {errors.title && <p className={styles.error}>{errors.title}</p>}
        <label>Description<textarea value={description} onChange={(event) => setDescription(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.description)} /></label>
        {errors.description && <p className={styles.error}>{errors.description}</p>}
        <div className={styles.row}>
          <label>Due date<input type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} disabled={saving} aria-invalid={Boolean(errors.dueDate)} /></label>
          <label>Status<select value={status} onChange={(event) => setStatus(event.target.value as TaskStatus)} disabled={saving} aria-invalid={Boolean(errors.status)}>{statuses.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
        </div>
        {errors.dueDate && <p className={styles.error}>{errors.dueDate}</p>}
        {errors.status && <p className={styles.error}>{errors.status}</p>}
        <button className={styles.primary} disabled={saving}>{saving ? "Saving..." : "Create task"}</button>
      </form>
      {toast && <p className={styles.toast} role="status">{toast}</p>}
    </div>
  );
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

function ErrorState() {
  return <section className={styles.state} role="alert"><strong>Cannot load tasks.</strong><p>Retry avoids stale browser data.</p></section>;
}

function Summary({ label, count }: { label: string; count: number }) {
  return <article className={styles.card}><p>{label}</p><strong>{count}</strong></article>;
}

function Column({ status, label, tasks }: { status: TaskStatus; label: string; tasks: Task[] }) {
  return (
    <section className={styles.column}>
      <h2><span className={styles[status]} />{label}<b>{tasks.length}</b></h2>
      {tasks.length ? tasks.map((task) => <TaskCard key={task.id} task={task} />) : <p className={styles.empty}>No tasks yet.</p>}
    </section>
  );
}

function TaskCard({ task }: { task: Task }) {
  return (
    <article className={styles.task}>
      <h3>{task.title}</h3>
      {task.description && <p>{task.description}</p>}
      {task.due_date && <time dateTime={task.due_date}>{task.due_date}</time>}
      <span className={`${styles.pill} ${styles[task.status]}`}>{task.status}</span>
    </article>
  );
}
