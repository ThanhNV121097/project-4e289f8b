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

export type TasksListResponse = {
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

const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

export async function fetchTasks(): Promise<TasksListResponse> {
  const tasks: Task[] = [];
  let cursor: string | null = null;
  let guard = 0;

  do {
    const params = new URLSearchParams({ limit: "200" });
    if (cursor) params.set("cursor", cursor);
    const response = await fetch(`${apiBase}/api/v1/tasks?${params}`, { headers: { Accept: "application/json" }, cache: "no-store" });
    if (!response.ok) throw await toApiError(response);
    const page = (await response.json()) as TasksListResponse;
    tasks.push(...page.tasks);
    cursor = page.next_cursor;
    guard += 1;
    if (guard > 20) throw new Error("Task list pagination exceeded safety limit.");
    if (page.has_more && !page.next_cursor) throw new Error("Task list pagination missing cursor.");
  } while (cursor);

  return { tasks, next_cursor: null, has_more: false };
}

export async function createTask(request: TaskCreateRequest): Promise<Task> {
  const response = await fetch(`${apiBase}/api/v1/tasks`, {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  if (!response.ok) throw await toApiError(response);
  return (await response.json()) as Task;
}

async function toApiError(response: Response): Promise<Error> {
  try {
    const body = (await response.json()) as ApiErrorResponse;
    return new Error(body.error.message);
  } catch {
    return new Error("Cannot reach task API. Retry to avoid stale board data.");
  }
}
