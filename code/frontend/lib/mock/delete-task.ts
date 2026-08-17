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

export const deleteTaskListResponse: TasksListResponse = {
  tasks: [
    {
      id: "550e8400-e29b-41d4-a716-446655440000",
      title: "Pay invoice",
      description: "Due before Friday",
      status: "todo",
      due_date: "2026-08-31",
      created_at: "2026-08-17T10:00:00Z",
      updated_at: "2026-08-17T10:00:00Z",
    },
    {
      id: "660e8400-e29b-41d4-a716-446655440001",
      title: "Draft board notes",
      description: "Mock upstream failure path keeps this saved task visible.",
      status: "doing",
      due_date: null,
      created_at: "2026-08-17T10:04:00Z",
      updated_at: "2026-08-17T10:04:00Z",
    },
    {
      id: "770e8400-e29b-41d4-a716-446655440002",
      title: "Archive old reminder",
      description: "Only done item, used to verify quiet empty state after delete.",
      status: "done",
      due_date: "2026-09-02",
      created_at: "2026-08-17T10:09:00Z",
      updated_at: "2026-08-17T10:09:00Z",
    },
  ],
  next_cursor: null,
  has_more: false,
};

export const deleteTaskErrorResponse: ApiErrorResponse = {
  error: {
    code: "UNAVAILABLE",
    message: "Cannot load tasks from the API. Retry to avoid stale board data.",
    details: [],
    request_id: "01HXDELETE000000000000000",
  },
};

export const deleteTaskNotFoundError: ApiErrorResponse = {
  error: {
    code: "NOT_FOUND",
    message: "Task was already deleted. Board refreshed from the API.",
    details: [],
    request_id: "01HXDELETE000000000000001",
  },
};

export const deleteTaskFailureError: ApiErrorResponse = {
  error: {
    code: "UNAVAILABLE",
    message: "Delete failed. Last confirmed saved task remains visible.",
    details: [],
    request_id: "01HXDELETE000000000000002",
  },
};
