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

export const tasksListResponse: TasksListResponse = {
  tasks: [
    {
      id: "550e8400-e29b-41d4-a716-446655440000",
      title: "Pay invoice",
      description: "Due before Friday",
      status: "todo",
      due_date: "2026-09-01",
      created_at: "2026-08-17T10:00:00Z",
      updated_at: "2026-08-17T10:00:00Z",
    },
    {
      id: "550e8400-e29b-41d4-a716-446655440001",
      title: "Draft agenda",
      description: null,
      status: "doing",
      due_date: null,
      created_at: "2026-08-17T10:05:00Z",
      updated_at: "2026-08-17T10:05:00Z",
    },
    {
      id: "550e8400-e29b-41d4-a716-446655440002",
      title: "Archive receipt",
      description: "Confirm records match bank export.",
      status: "done",
      due_date: "2026-09-03",
      created_at: "2026-08-17T10:10:00Z",
      updated_at: "2026-08-17T10:10:00Z",
    },
  ],
  next_cursor: null,
  has_more: false,
};

export const apiErrorResponse: ApiErrorResponse = {
  error: {
    code: "UNAVAILABLE",
    message: "Cannot load tasks from API. Retry to avoid stale board data.",
    details: [],
    request_id: "01HX0000000000000000000000",
  },
};
