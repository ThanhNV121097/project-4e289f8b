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

export type TasksResponse = {
  tasks: Task[];
  next_cursor: string | null;
  has_more: boolean;
};

export type ApiError = {
  error: {
    code: "BAD_REQUEST" | "VALIDATION_FAILED" | "NOT_FOUND" | "RATE_LIMITED" | "INTERNAL" | "UNAVAILABLE";
    message: string;
    details: { field: string; code: string; message: string }[];
    request_id: string;
  };
};

type PatchTask = Partial<Pick<Task, "title" | "description" | "status" | "due_date">>;

const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

async function parseError(response: Response): Promise<never> {
  let error: ApiError;
  try {
    error = await response.json();
  } catch {
    error = { error: { code: "INTERNAL", message: "Unexpected API response.", details: [], request_id: response.headers.get("X-Request-Id") ?? "" } };
  }
  throw error;
}

export async function loadTasks(): Promise<TasksResponse> {
  const response = await fetch(`${baseUrl}/api/v1/tasks`, { headers: { Accept: "application/json" }, cache: "no-store" });
  if (!response.ok) return parseError(response);
  return response.json();
}

export async function updateTask(taskId: string, patch: PatchTask): Promise<Task> {
  const response = await fetch(`${baseUrl}/api/v1/tasks/${taskId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(patch),
  });
  if (!response.ok) return parseError(response);
  return response.json();
}
