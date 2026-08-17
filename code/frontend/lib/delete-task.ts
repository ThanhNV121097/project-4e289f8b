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
    code: "BAD_REQUEST" | "NOT_FOUND" | "RATE_LIMITED" | "INTERNAL" | "UNAVAILABLE";
    message: string;
    details: { field: string; code: string; message: string }[];
    request_id: string;
  };
};

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

export async function listTasks(): Promise<TasksListResponse> {
  const response = await fetch(`${API_BASE}/api/v1/tasks`, { headers: { Accept: "application/json" }, cache: "no-store" });
  if (!response.ok) throw await readApiError(response);
  return response.json() as Promise<TasksListResponse>;
}

export async function deleteTask(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/api/v1/tasks/${id}`, { method: "DELETE" });
  if (response.status === 204) return;
  throw await readApiError(response);
}

async function readApiError(response: Response): Promise<ApiErrorResponse> {
  try {
    return (await response.json()) as ApiErrorResponse;
  } catch {
    return {
      error: {
        code: "UNAVAILABLE",
        message: "API request failed. Last confirmed saved task remains visible.",
        details: [],
        request_id: response.headers.get("X-Request-Id") ?? "unknown",
      },
    };
  }
}
