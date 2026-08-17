"use client";

import { useEffect, useMemo, useState } from "react";
import { apiErrorResponse, tasksListResponse, type Task, type TaskStatus } from "../lib/mock/view-tasks-by-status";
import styles from "./ViewTasksByStatus.module.css";

const statuses: Array<{ key: TaskStatus; title: string }> = [
  { key: "todo", title: "Todo" },
  { key: "doing", title: "Doing" },
  { key: "done", title: "Done" },
];

type LoadState = "loading" | "ready" | "error";

export default function ViewTasksByStatus() {
  const [state, setState] = useState<LoadState>("loading");
  const [tasks, setTasks] = useState<Task[]>([]);

  useEffect(() => {
    loadTasks();
  }, []);

  function loadTasks() {
    setState("loading");
    window.setTimeout(() => {
      const valid = tasksListResponse.tasks.every((task) => isTaskStatus(task.status));
      setTasks(valid ? tasksListResponse.tasks : []);
      setState(valid ? "ready" : "error");
    }, 300);
  }

  const groups = useMemo(() => groupTasks(tasks), [tasks]);

  return (
    <div className={styles.shell}>
      <Topbar />

      <section className={styles.hero} aria-labelledby="task-board-title">
        <p className={styles.eyebrow}>One board · API source of truth</p>
        <h1 id="task-board-title">Task board</h1>
        <p>See every persisted task grouped by Todo, Doing, and Done. Browser storage is not used for task truth.</p>
      </section>

      <section className={styles.summary} aria-label="Task totals">
        {statuses.map((status) => (
          <article className={styles.summaryCard} key={status.key}>
            <span>{status.title}</span>
            <strong>{groups[status.key].length}</strong>
          </article>
        ))}
      </section>

      {state === "loading" ? <LoadingState /> : null}
      {state === "error" ? <ErrorState onRetry={loadTasks} /> : null}

      <section className={state === "loading" ? styles.boardLoading : styles.board} aria-label="Task board">
        {statuses.map((status) => (
          <BoardColumn key={status.key} status={status} tasks={groups[status.key]} />
        ))}
      </section>
    </div>
  );
}

function Topbar() {
  return (
    <header className={styles.topbar}>
      <a className={styles.brand} href="#task-board-title" aria-label="Task board home">
        <span className={styles.mark} aria-hidden="true" />
        <span>Task board</span>
      </a>
      <nav className={styles.nav} aria-label="Page navigation">
        <a href="#task-board-title">Board</a>
        <a href="#task-board-title">Create task</a>
        <a href="#task-board-title">States</a>
      </nav>
    </header>
  );
}

function groupTasks(tasks: Task[]): Record<TaskStatus, Task[]> {
  return {
    todo: tasks.filter((task) => task.status === "todo"),
    doing: tasks.filter((task) => task.status === "doing"),
    done: tasks.filter((task) => task.status === "done"),
  };
}

function isTaskStatus(status: string): status is TaskStatus {
  return status === "todo" || status === "doing" || status === "done";
}

function LoadingState() {
  return (
    <section className={styles.panel} aria-live="polite">
      <strong>Loading tasks from API.</strong>
      <span className={styles.skeleton} />
      <span className={styles.skeletonShort} />
    </section>
  );
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <section className={styles.panel} role="alert">
      <strong>Cannot load tasks.</strong>
      <p>{apiErrorResponse.error.message}</p>
      <button className={styles.button} onClick={onRetry} type="button">Retry</button>
    </section>
  );
}
function BoardColumn({ status, tasks }: { status: { key: TaskStatus; title: string }; tasks: Task[] }) {
  return (
    <section className={styles.column} aria-labelledby={`${status.key}-heading`}>
      <header className={styles.columnHeader}>
        <span className={`${styles.dot} ${styles[status.key]}`} />
        <h2 id={`${status.key}-heading`}>{status.title}</h2>
        <span className={styles.count}>{tasks.length}</span>
      </header>
      <div className={styles.cards}>
        {tasks.length === 0 ? <p className={styles.empty}>No {status.key} tasks.</p> : tasks.map((task) => <TaskCard key={task.id} task={task} />)}
      </div>
    </section>
  );
}

function TaskCard({ task }: { task: Task }) {
  const moveText = task.status === "done" ? "Move left" : "Move right";

  return (
    <article className={styles.card}>
      <h3>{task.title}</h3>
      {task.description ? <p>{task.description}</p> : null}
      <div className={styles.meta}>
        {task.due_date ? <span>Due {task.due_date}</span> : null}
        <span className={`${styles.pill} ${styles[task.status]}`}>{task.status}</span>
      </div>
      <div className={styles.actions} aria-label={`Actions for ${task.title}`}>
        <button className={styles.mini} type="button">{moveText}</button>
        <button className={styles.mini} type="button">Edit</button>
        <button className={styles.danger} type="button">Delete</button>
      </div>
    </article>
  );
}
