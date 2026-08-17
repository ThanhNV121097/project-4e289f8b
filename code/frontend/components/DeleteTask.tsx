"use client";

import { useMemo, useState } from "react";
import {
  deleteTaskEmptyResponse,
  deleteTaskErrorResponse,
  deleteTaskListResponse,
  deleteTaskNotFoundError,
  type Task,
  type TaskStatus,
} from "../lib/mock/delete-task";
import styles from "./DeleteTask.module.css";

const statuses: TaskStatus[] = ["todo", "doing", "done"];
const labels: Record<TaskStatus, string> = { todo: "Todo", doing: "Doing", done: "Done" };
type LoadState = "loading" | "error" | "ready" | "empty";

export default function DeleteTask() {
  const [tasks, setTasks] = useState<Task[]>(deleteTaskListResponse.tasks);
  const [loadState, setLoadState] = useState<LoadState>("ready");
  const [target, setTarget] = useState<Task | null>(null);
  const [message, setMessage] = useState("");
  const counts = useMemo(() => countByStatus(tasks), [tasks]);

  function retryLoad() {
    setLoadState("loading");
    window.setTimeout(() => setLoadState("ready"), 160);
  }

  function confirmDelete() {
    if (!target) return;
    if (!tasks.some((task) => task.id === target.id)) {
      setMessage(deleteTaskNotFoundError.error.message);
      setTarget(null);
      return;
    }
    setTasks(tasks.filter((task) => task.id !== target.id));
    setMessage(`Deleted “${target.title}” through DELETE /api/v1/tasks/${target.id}.`);
    setTarget(null);
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
      {loadState === "empty" ? <Board tasks={deleteTaskEmptyResponse.tasks} counts={countByStatus([])} onDelete={setTarget} /> : null}
      {loadState === "ready" ? <Board tasks={tasks} counts={counts} onDelete={setTarget} /> : null}
      {target ? <Confirm task={target} onCancel={() => setTarget(null)} onConfirm={confirmDelete} /> : null}
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
      <div className={styles.meta}><span>Due {task.due_date ?? "None"}</span><span className={`${styles.pill} ${styles[task.status]}`}>{labels[task.status]}</span></div>
      <div className={styles.actions}>
        <button type="button" disabled>Move left</button><button type="button" disabled>Move right</button><button type="button" disabled>Edit</button>
        <button className={styles.danger} type="button" onClick={() => onDelete(task)}>Delete</button>
      </div>
    </article>
  );
}

function Confirm({ task, onCancel, onConfirm }: { task: Task; onCancel: () => void; onConfirm: () => void }) {
  return (
    <div className={styles.backdrop} role="presentation" onClick={onCancel}>
      <section className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="confirm-delete-title" onClick={(event) => event.stopPropagation()}>
        <h2 id="confirm-delete-title">Delete task?</h2>
        <p>“{task.title}” will be removed from Postgres through REST API. This cannot be undone.</p>
        <div className={styles.modalActions}>
          <button type="button" onClick={onCancel}>Cancel</button>
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
  return <section className={styles.state} role="alert"><strong>Cannot load tasks.</strong><p>{deleteTaskErrorResponse.error.message}</p><button type="button" onClick={onRetry}>Retry</button></section>;
}

function EmptyState({ status }: { status: TaskStatus }) {
  return <div className={styles.empty}><span aria-hidden="true">□</span><p>No {status} tasks.</p></div>;
}

function countByStatus(tasks: Task[]) {
  return statuses.reduce<Record<TaskStatus, number>>((counts, status) => ({ ...counts, [status]: tasks.filter((task) => task.status === status).length }), { todo: 0, doing: 0, done: 0 });
}
