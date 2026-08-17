export type TaskStatus = "todo" | "doing" | "done";

export type Task = {
  id: string;
  title: string;
  description: string | null;
  status: TaskStatus;
  due_date: string | null;
  created_at: string;
  updated_at: string;
};

export type TaskListResponse = {
  tasks: Task[];
  next_cursor: string | null;
  has_more: boolean;
};

export type TaskCreateRequest = {
  title: string;
  description?: string | null;
  status?: TaskStatus;
  due_date?: string | null;
};

export type ApiErrorResponse = {
  error: {
    code: "BAD_REQUEST" | "VALIDATION_FAILED" | "RATE_LIMITED" | "INTERNAL" | "UNAVAILABLE";
    message: string;
    details: Array<{ field: string; code: string; message: string }>;
    request_id: string;
  };
};

const now = "2026-08-17T10:00:00Z";
let savedTasks: Task[] = [
  {
    id: "550e8400-e29b-41d4-a716-446655440000",
    title: "Pay invoice",
    description: "Due before Friday",
    status: "todo",
    due_date: "2026-08-31",
    created_at: now,
    updated_at: now,
  },
];

export async function listTasks(): Promise<TaskListResponse> {
  await pause();
  return { tasks: [...savedTasks], next_cursor: null, has_more: false };
}

export async function createTask(request: TaskCreateRequest): Promise<Task> {
  await pause();
  const saved: Task = {
    id: crypto.randomUUID(),
    title: request.title.trim(),
    description: request.description?.trim() || null,
    status: request.status ?? "todo",
    due_date: request.due_date || null,
    created_at: now,
    updated_at: now,
  };
  savedTasks = [...savedTasks, saved];
  return saved;
}

function pause() {
  return new Promise((resolve) => setTimeout(resolve, 120));
}
