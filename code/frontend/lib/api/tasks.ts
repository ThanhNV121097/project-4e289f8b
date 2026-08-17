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

export type ApiErrorResponse = {
  error: {
    code: "BAD_REQUEST" | "RATE_LIMITED" | "INTERNAL" | "UNAVAILABLE";
    message: string;
    details: Array<{ field: string; code: string; message: string }>;
    request_id: string;
  };
};

// The scaffold's contract, not a choice: local compose bakes
// NEXT_PUBLIC_API_URL=http://localhost:8080 (straight to the backend, which
// mounts /v1/... with no prefix); production bakes nothing, the fallback
// "/api" applies, and the edge proxy strips it back off before forwarding.
// The old name here (NEXT_PUBLIC_API_BASE_URL) was read by nobody.
const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "/api";

export async function fetchTasks(): Promise<TasksListResponse> {
  const tasks: Task[] = [];
  let cursor: string | null = null;
  let guard = 0;

  do {
    const params = new URLSearchParams({ limit: "200" });
    if (cursor) params.set("cursor", cursor);
    const response = await fetch(`${apiBase}/v1/tasks?${params}`, { headers: { Accept: "application/json" }, cache: "no-store" });
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

async function toApiError(response: Response): Promise<Error> {
  try {
    const body = (await response.json()) as ApiErrorResponse;
    return new Error(body.error.message);
  } catch {
    return new Error("Cannot load tasks from API. Retry to avoid stale board data.");
  }
}
