"use client";

import { useEffect, useMemo, useState } from "react";
import {
  createTask,
  deleteTask,
  fetchTasks,
  updateTask,
  type Task,
  type TaskInput,
  type TaskStatus,
} from "../lib/api/tasks";
import CreateTaskForm from "./board/CreateTaskForm";
import { DeleteTaskDialog, EditTaskDialog } from "./board/Dialogs";
import boardStyles from "./board/board.module.css";
import styles from "./ViewTasksByStatus.module.css";

const statuses: Array<{ key: TaskStatus; title: string }> = [
  { key: "todo", title: "Todo" },
  { key: "doing", title: "Doing" },
  { key: "done", title: "Done" },
];

const statusTitles: Record<TaskStatus, string> = {
  todo: "Todo",
  doing: "Doing",
  done: "Done",
};

const statusOrder: TaskStatus[] = ["todo", "doing", "done"];

type LoadState = "loading" | "ready" | "error";
type Toast = { kind: "success" | "error"; text: string };

export default function ViewTasksByStatus() {
  const [state, setState] = useState<LoadState>("loading");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [errorMessage, setErrorMessage] = useState("Cannot load tasks from API. Retry to avoid stale board data.");
  const [toast, setToast] = useState<Toast | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [editing, setEditing] = useState<Task | null>(null);
  const [deleting, setDeleting] = useState<Task | null>(null);

  useEffect(() => {
    loadTasks();
  }, []);

  useEffect(() => {
    if (!toast) return;
    const timer = setTimeout(() => setToast(null), 4500);
    return () => clearTimeout(timer);
  }, [toast]);

  function loadTasks() {
    setState("loading");
    fetchTasks()
      .then((response) => {
        const valid = response.tasks.every((task) => isTaskStatus(task.status));
        setTasks(valid ? response.tasks : []);
        setErrorMessage(valid ? "Cannot load tasks from API. Retry to avoid stale board data." : "API returned an invalid task status.");
        setState(valid ? "ready" : "error");
      })
      .catch((error: Error) => {
        setTasks([]);
        setErrorMessage(error.message);
        setState("error");
      });
  }

  // After a mutation: re-read from the API so the board shows what the
  // database holds, but never blank the board on a refresh failure — the last
  // confirmed saved tasks stay visible and the failure is reported instead.
  async function quietReload() {
    try {
      const response = await fetchTasks();
      setTasks(response.tasks.filter((task) => isTaskStatus(task.status)));
      setState("ready");
    } catch (error) {
      setToast({ kind: "error", text: (error as Error).message });
    }
  }

  async function mutate(key: string, action: () => Promise<void>, success: string): Promise<boolean> {
    setBusy(key);
    try {
      await action();
      await quietReload();
      setToast({ kind: "success", text: success });
      return true;
    } catch (error) {
      setToast({ kind: "error", text: (error as Error).message });
      return false;
    } finally {
      setBusy(null);
    }
  }

  function handleCreate(input: TaskInput) {
    return mutate("create", async () => void (await createTask(input)), "Task created through REST API.");
  }

  function handleMove(task: Task, direction: -1 | 1) {
    const target = statusOrder[statusOrder.indexOf(task.status) + direction];
    if (!target) return;
    void mutate(task.id, async () => void (await updateTask(task.id, { status: target })), `Task moved to ${statusTitles[target]} through REST API.`);
  }

  async function handleSaveEdit(patch: { title: string; description: string | null; due_date: string | null }) {
    if (!editing) return;
    const saved = await mutate(editing.id, async () => void (await updateTask(editing.id, patch)), "Task updated through REST API.");
    if (saved) setEditing(null);
  }

  async function handleConfirmDelete() {
    if (!deleting) return;
    const removed = await mutate(deleting.id, () => deleteTask(deleting.id), "Task deleted through REST API.");
    if (removed) setDeleting(null);
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
      {state === "error" ? <ErrorState message={errorMessage} onRetry={loadTasks} /> : null}
      <section className={state === "loading" ? styles.boardLoading : styles.board} aria-label="Task board">
        {statuses.map((status) => (
          <BoardColumn
            key={status.key}
            status={status}
            tasks={groups[status.key]}
            busy={busy}
            onMove={handleMove}
            onEdit={setEditing}
            onDelete={setDeleting}
          />
        ))}
      </section>
      <CreateTaskForm busy={busy === "create"} onCreate={handleCreate} />
      {editing ? <EditTaskDialog task={editing} busy={busy === editing.id} onCancel={() => setEditing(null)} onSave={handleSaveEdit} /> : null}
      {deleting ? <DeleteTaskDialog task={deleting} busy={busy === deleting.id} onCancel={() => setDeleting(null)} onConfirm={handleConfirmDelete} /> : null}
      {toast ? (
        <p className={toast.kind === "error" ? `${boardStyles.toast} ${boardStyles.toastError}` : boardStyles.toast} role="status" aria-live="polite">
          {toast.text}
        </p>
      ) : null}
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
        <a href="#create-task-title">Create task</a>
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

function formatDueDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", year: "numeric" }).format(new Date(`${value}T00:00:00`));
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

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <section className={styles.panel} role="alert">
      <strong>Cannot load tasks.</strong>
      <p>{message}</p>
      <button className={styles.button} onClick={onRetry} type="button">Retry</button>
    </section>
  );
}

type CardHandlers = {
  busy: string | null;
  onMove: (task: Task, direction: -1 | 1) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
};

function BoardColumn({ status, tasks, ...handlers }: { status: { key: TaskStatus; title: string }; tasks: Task[] } & CardHandlers) {
  return (
    <section className={styles.column} aria-labelledby={`${status.key}-heading`}>
      <header className={styles.columnHeader}>
        <span className={`${styles.dot} ${styles[status.key]}`} />
        <h2 id={`${status.key}-heading`}>{status.title}</h2>
        <span className={styles.count}>{tasks.length}</span>
      </header>
      <div className={styles.cards}>
        {tasks.length === 0 ? <p className={styles.empty}>No {status.key} tasks.</p> : tasks.map((task) => <TaskCard key={task.id} task={task} {...handlers} />)}
      </div>
    </section>
  );
}

function TaskCard({ task, busy, onMove, onEdit, onDelete }: { task: Task } & CardHandlers) {
  const disabled = busy !== null;

  return (
    <article className={styles.card}>
      <h3>{task.title}</h3>
      {task.description ? <p>{task.description}</p> : null}
      <div className={styles.meta}>
        {task.due_date ? <span>{formatDueDate(task.due_date)}</span> : null}
        <span className={`${styles.pill} ${styles[task.status]}`}>{statusTitles[task.status]}</span>
      </div>
      <div className={styles.actions} aria-label={`Actions for ${task.title}`}>
        {task.status !== "todo" ? (
          <button className={styles.mini} type="button" disabled={disabled} onClick={() => onMove(task, -1)}>Move left</button>
        ) : null}
        {task.status !== "done" ? (
          <button className={styles.mini} type="button" disabled={disabled} onClick={() => onMove(task, 1)}>Move right</button>
        ) : null}
        <button className={styles.mini} type="button" disabled={disabled} onClick={() => onEdit(task)}>Edit</button>
        <button className={styles.danger} type="button" disabled={disabled} onClick={() => onDelete(task)}>Delete</button>
      </div>
    </article>
  );
}
