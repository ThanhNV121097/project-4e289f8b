"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { deleteTask, listTasks, type ApiErrorResponse, type Task, type TaskStatus } from "../lib/delete-task";
import styles from "./DeleteTask.module.css";

const statuses: TaskStatus[] = ["todo", "doing", "done"];
const labels: Record<TaskStatus, string> = { todo: "Todo", doing: "Doing", done: "Done" };
type LoadState = "loading" | "error" | "ready";

export default function DeleteTask() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [target, setTarget] = useState<Task | null>(null);
  const [message, setMessage] = useState("");
  const restoreFocusRef = useRef<HTMLElement | null>(null);
  const counts = useMemo(() => countByStatus(tasks), [tasks]);

  useEffect(() => { void retryLoad(); }, []);

  async function retryLoad() {
    setLoadState("loading");
    try {
      const response = await listTasks();
      setTasks(response.tasks);
      setLoadState("ready");
    } catch (error) {
      setMessage(apiMessage(error, "Cannot load tasks from the API. Retry to avoid stale board data."));
      setLoadState("error");
    }
  }

  function requestDelete(task: Task) {
    restoreFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    setTarget(task);
  }

  function closeConfirm() {
    setTarget(null);
    window.setTimeout(() => restoreFocusRef.current?.focus(), 0);
  }

  async function confirmDelete() {
    if (!target) return;
    try {
      await deleteTask(target.id);
      setTasks(tasks.filter((task) => task.id !== target.id));
      setMessage(`Deleted “${target.title}” through DELETE /api/v1/tasks/${target.id}.`);
    } catch (error) {
      setMessage(apiMessage(error, "Delete failed. Last confirmed saved task remains visible."));
      if ((error as ApiErrorResponse).error?.code === "NOT_FOUND") void retryLoad();
    }
    closeConfirm();
  }

  return (
    <section className={styles.wrap} aria-labelledby="delete-task-title">
      <div className={styles.hero}>
        <div>
          <p className={styles.eyebrow}>Delete task story</p>
          <h1 id="delete-task-title">Remove unwanted tasks from one board</h1>
          <p>Confirm delete first. Successful delete removes card and updates counts from saved API state.</p>
        </div>
      </div>

      {message ? <p className={styles.toast} role="status">{message}</p> : null}
      {loadState === "loading" ? <LoadingState /> : null}
      {loadState === "error" ? <ErrorState onRetry={retryLoad} /> : null}
      {loadState === "ready" ? <Board tasks={tasks} counts={counts} onDelete={requestDelete} /> : null}
      {target ? <Confirm task={target} onCancel={closeConfirm} onConfirm={confirmDelete} /> : null}
    </section>
  );
}

function Board({ tasks, counts, onDelete }: { tasks: Task[]; counts: Record<TaskStatus, number>; onDelete: (task: Task) => void }) {
  return (
    <>
      <section className={styles.summary} aria-label="Task totals">
        {statuses.map((status) => <article key={status}><span>{labels[status]}</span><strong>{counts[status]}</strong></article>)}
      </section>
      <section className={styles.board} aria-label="Task board">
        {statuses.map((status) => <Column key={status} status={status} tasks={tasks.filter((task) => task.status === status)} onDelete={onDelete} />)}
      </section>
    </>
  );
}

function Column({ status, tasks, onDelete }: { status: TaskStatus; tasks: Task[]; onDelete: (task: Task) => void }) {
  return (
    <section className={styles.column}>
      <header><h2><span className={`${styles.dot} ${styles[status]}`} />{labels[status]}</h2><span>{tasks.length}</span></header>
      <div className={styles.cards}>{tasks.length ? tasks.map((task) => <TaskCard key={task.id} task={task} onDelete={onDelete} />) : <EmptyState status={status} />}</div>
    </section>
  );
}

function TaskCard({ task, onDelete }: { task: Task; onDelete: (task: Task) => void }) {
  return (
    <article className={styles.card}>
      <h3>{task.title}</h3>
      <p>{task.description ?? "No description"}</p>
      <div className={styles.meta}><span>Due {formatDueDate(task.due_date)}</span><span className={`${styles.pill} ${styles[task.status]}`}>{labels[task.status]}</span></div>
      <div className={styles.actions}>
        <button type="button" disabled>Move left</button><button type="button" disabled>Move right</button><button type="button" disabled>Edit</button>
        <button className={styles.danger} type="button" onClick={() => onDelete(task)}>Delete</button>
      </div>
    </article>
  );
}

function Confirm({ task, onCancel, onConfirm }: { task: Task; onCancel: () => void; onConfirm: () => void }) {
  const modalRef = useRef<HTMLElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    cancelRef.current?.focus();
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onCancel();
      if (event.key === "Tab") trapFocus(event, modalRef.current);
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onCancel]);

  return (
    <div className={styles.backdrop} role="presentation" onClick={onCancel}>
      <section ref={modalRef} className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="confirm-delete-title" onClick={(event) => event.stopPropagation()}>
        <h2 id="confirm-delete-title">Delete task?</h2>
        <p>“{task.title}” will be removed from Postgres through REST API. This cannot be undone.</p>
        <div className={styles.modalActions}>
          <button ref={cancelRef} type="button" onClick={onCancel}>Cancel</button>
          <button className={styles.primaryDanger} type="button" onClick={onConfirm}>Confirm delete</button>
        </div>
      </section>
    </div>
  );
}

function LoadingState() {
  return <section className={styles.state} aria-live="polite"><strong>Loading tasks from API.</strong><span /><span /></section>;
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return <section className={styles.state} role="alert"><strong>Cannot load tasks.</strong><p>API data unavailable.</p><button type="button" onClick={onRetry}>Retry</button></section>;
}

function EmptyState({ status }: { status: TaskStatus }) {
  return <div className={styles.empty}><span aria-hidden="true">□</span><p>No {status} tasks.</p></div>;
}

function countByStatus(tasks: Task[]) {
  return statuses.reduce<Record<TaskStatus, number>>((counts, status) => ({ ...counts, [status]: tasks.filter((task) => task.status === status).length }), { todo: 0, doing: 0, done: 0 });
}

function formatDueDate(dueDate: string | null) {
  if (!dueDate) return "None";
  const [year, month, day] = dueDate.split("-").map(Number);
  return new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", year: "numeric" }).format(new Date(year, month - 1, day));
}

function apiMessage(error: unknown, fallback: string) {
  return (error as ApiErrorResponse).error?.message ?? fallback;
}

function trapFocus(event: KeyboardEvent, modal: HTMLElement | null) {
  if (!modal) return;
  const focusable = Array.from(modal.querySelectorAll<HTMLElement>("button:not(:disabled), [href], input, select, textarea, [tabindex]:not([tabindex='-1'])"));
  const first = focusable[0];
  const last = focusable[focusable.length - 1];
  if (!first || !last) return;
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}
