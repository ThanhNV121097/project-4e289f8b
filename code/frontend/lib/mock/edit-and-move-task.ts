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

export const tasksResponse: TasksResponse = {
  tasks: [
    {
      id: "550e8400-e29b-41d4-a716-446655440000",
      title: "Draft",
      description: null,
      status: "todo",
      due_date: null,
      created_at: "2026-08-17T10:00:00Z",
      updated_at: "2026-08-17T10:00:00Z",
    },
    {
      id: "550e8400-e29b-41d4-a716-446655440001",
      title: "Call warehouse",
      description: "Confirm pickup window",
      status: "doing",
      due_date: "2026-08-31",
      created_at: "2026-08-17T10:03:00Z",
      updated_at: "2026-08-17T10:04:00Z",
    },
    {
      id: "550e8400-e29b-41d4-a716-446655440002",
      title: "Send receipt",
      description: "Old",
      status: "done",
      due_date: "2026-09-02",
      created_at: "2026-08-17T10:06:00Z",
      updated_at: "2026-08-17T10:07:00Z",
    },
  ],
  next_cursor: null,
  has_more: false,
};

export const loadError: ApiError = {
  error: {
    code: "UNAVAILABLE",
    message: "Could not load tasks.",
    details: [],
    request_id: "01HX0000000000000000000000",
  },
};
