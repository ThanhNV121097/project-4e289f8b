"use client";

import { FormEvent, KeyboardEvent, ReactNode, useEffect, useMemo, useRef, useState } from "react";
import { loadTasks, updateTask, Task, TaskStatus } from "../lib/edit-and-move-task";
import styles from "./EditAndMoveTask.module.css";

const statuses: { key: TaskStatus; label: string }[] = [
  { key: "todo", label: "Todo" },
  { key: "doing", label: "Doing" },
  { key: "done", label: "Done" },
];

type Errors = Partial<Record<"title" | "description" | "due_date" | "status", string>>;

function nextStatus(status: TaskStatus, delta: number) {
  const index = statuses.findIndex((item) => item.key === status);
  return statuses[Math.max(0, Math.min(statuses.length - 1, index + delta))].key;
}

function fieldErrors(error: unknown): Errors {
  const details = (error as { error?: { details?: { field: string; message: string }[] } }).error?.details ?? [];
  return Object.fromEntries(details.map((detail) => [detail.field, detail.message])) as Errors;
}

export default function EditAndMoveTask() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [editing, setEditing] = useState<Task | null>(null);
  const [errors, setErrors] = useState<Errors>({});
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const opener = useRef<HTMLButtonElement | null>(null);
  const totals = useMemo(() => statuses.map((s) => ({ ...s, count: tasks.filter((t) => t.status === s.key).length })), [tasks]);

  async function refresh() {
    setLoading(true);
    setLoadError(false);
    try { setTasks((await loadTasks()).tasks); } catch { setLoadError(true); } finally { setLoading(false); }
  }

  useEffect(() => { void refresh(); }, []);

  function closeModal() { setEditing(null); setErrors({}); setSaveError(null); opener.current?.focus(); }

  async function saveTask(task: Task) {
    setErrors({}); setSaveError(null);
    try {
      const saved = await updateTask(task.id, { title: task.title, description: task.description, status: task.status, due_date: task.due_date });
      setTasks((current) => current.map((item) => (item.id === saved.id ? saved : item)));
      closeModal();
    } catch (error) {
      setErrors(fieldErrors(error));
      setSaveError((error as { error?: { code?: string; message?: string } }).error?.code === "NOT_FOUND" ? "Task not found. Board refreshed from API." : (error as { error?: { message?: string } }).error?.message ?? "Could not save task.");
      if ((error as { error?: { code?: string } }).error?.code === "NOT_FOUND") void refresh();
    }
  }

  function moveTask(task: Task, delta: number) { void saveTask({ ...task, status: nextStatus(task.status, delta) }); }

  return <section className={styles.shell} aria-label="Edit and move tasks">
    {loading && <StatePanel kind="loading" onRetry={refresh} />}
    {loadError && <StatePanel kind="error" onRetry={refresh} />}
    {saveError && <div className={styles.state} role="alert"><strong>{saveError}</strong></div>}
    <div className={styles.summary} aria-label="Task totals">{totals.map((item) => <div className={styles.summaryCard} key={item.key}><span>{item.label}</span><strong>{item.count}</strong></div>)}</div>
    <div className={styles.board} aria-label="Task board">{statuses.map((status) => <Column key={status.key} status={status} tasks={tasks.filter((task) => task.status === status.key)} onEdit={(task, button) => { opener.current = button; setEditing(task); }} onMove={moveTask} />)}</div>
    {editing && <EditModal task={editing} errors={errors} onCancel={closeModal} onSave={saveTask} />}
  </section>;
}

function Column({ status, tasks, onEdit, onMove }: { status: { key: TaskStatus; label: string }; tasks: Task[]; onEdit: (task: Task, button: HTMLButtonElement) => void; onMove: (task: Task, delta: number) => void }) {
  return <section className={styles.column}><ColumnHeader status={status} count={tasks.length} /><div className={styles.cards}>{tasks.length ? tasks.map((task) => <TaskCard key={task.id} task={task} onEdit={onEdit} onMove={onMove} />) : <div className={styles.empty}><span aria-hidden="true">□</span>No {status.label.toLowerCase()} tasks.</div>}</div></section>;
}

function TaskCard({ task, onEdit, onMove }: { task: Task; onEdit: (task: Task, button: HTMLButtonElement) => void; onMove: (task: Task, delta: number) => void }) {
  return <article className={styles.card}><h3>{task.title}</h3><p>{task.description || "No description"}</p>{task.due_date && <span className={styles.meta}>Due {new Date(`${task.due_date}T00:00:00`).toLocaleDateString("en", { month: "short", day: "numeric", year: "numeric" })}</span>}<span className={`${styles.pill} ${styles[task.status]}`}>{statuses.find((item) => item.key === task.status)?.label}</span><div className={styles.actions}><button disabled={task.status === "todo"} onClick={() => onMove(task, -1)}>Move left</button><button disabled={task.status === "done"} onClick={() => onMove(task, 1)}>Move right</button><button onClick={(event) => onEdit(task, event.currentTarget)}>Edit</button><button disabled aria-disabled="true">Delete</button></div></article>;
}

function ColumnHeader({ status, count }: { status: { key: TaskStatus; label: string }; count: number }) { return <header className={styles.columnHeader}><span className={`${styles.dot} ${styles[status.key]}`} /> <h2>{status.label}</h2><span>{count}</span></header>; }
function StatePanel({ kind, onRetry }: { kind: "loading" | "error"; onRetry: () => void }) { if (kind === "error") return <div className={styles.state} role="alert"><strong>Could not load tasks.</strong><p>Retry avoids showing stale browser data.</p><button onClick={onRetry}>Retry</button></div>; return <div className={styles.state} aria-live="polite"><strong>Loading tasks from REST API.</strong><span /><span /></div>; }

function EditModal({ task, errors, onCancel, onSave }: { task: Task; errors: Errors; onCancel: () => void; onSave: (task: Task) => void }) {
  const [draft, setDraft] = useState(task);
  const modalRef = useRef<HTMLFormElement | null>(null);
  useEffect(() => { modalRef.current?.querySelector<HTMLElement>("input, select, textarea, button")?.focus(); }, []);
  function submit(event: FormEvent) { event.preventDefault(); void onSave(draft); }
  function trapFocus(event: KeyboardEvent<HTMLFormElement>) {
    if (event.key === "Escape") { event.preventDefault(); onCancel(); return; }
    if (event.key !== "Tab") return;
    const focusable = Array.from(modalRef.current?.querySelectorAll<HTMLElement>("button, input, select, textarea") ?? []).filter((item) => !item.hasAttribute("disabled"));
    const first = focusable[0]; const last = focusable[focusable.length - 1];
    if (!first || !last) return;
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
    if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
  }
  return <div className={styles.backdrop} onMouseDown={onCancel}><form ref={modalRef} className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="edit-title" onSubmit={submit} onKeyDown={trapFocus} onMouseDown={(event) => event.stopPropagation()}><header><h2 id="edit-title">Edit task</h2><button type="button" onClick={onCancel}>Close</button></header><Field id="edit-task-title" label="Title" error={errors.title}><input id="edit-task-title" value={draft.title} aria-describedby={errors.title ? "edit-task-title-error" : undefined} onChange={(e) => setDraft({ ...draft, title: e.target.value })} /></Field><Field id="edit-task-description" label="Description" error={errors.description}><textarea id="edit-task-description" value={draft.description ?? ""} aria-describedby={errors.description ? "edit-task-description-error" : undefined} onChange={(e) => setDraft({ ...draft, description: e.target.value })} /></Field><Field id="edit-task-due" label="Due date" error={errors.due_date}><input id="edit-task-due" type="date" value={draft.due_date ?? ""} aria-describedby={errors.due_date ? "edit-task-due-error" : undefined} onChange={(e) => setDraft({ ...draft, due_date: e.target.value || null })} /></Field><Field id="edit-task-status" label="Status" error={errors.status}><select id="edit-task-status" value={draft.status} aria-describedby={errors.status ? "edit-task-status-error" : undefined} onChange={(e) => setDraft({ ...draft, status: e.target.value as TaskStatus })}>{statuses.map((s) => <option value={s.key} key={s.key}>{s.label}</option>)}</select></Field><footer><button type="button" onClick={onCancel}>Cancel</button><button className={styles.primary} type="submit">Save changes</button></footer></form></div>;
}

function Field({ id, label, error, children }: { id: string; label: string; error?: string; children: ReactNode }) { return <label className={`${styles.field} ${error ? styles.invalid : ""}`} htmlFor={id}><span>{label}</span>{children}{error && <small id={`${id}-error`}>{error}</small>}</label>; }
