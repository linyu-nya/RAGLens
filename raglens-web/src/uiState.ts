export type ResourceState =
  | { kind: "loading"; message: string }
  | { kind: "error"; message: string }
  | { kind: "empty"; message: string }
  | { kind: "ready" };

export type ResourceStateInput = {
  loading?: boolean;
  error?: unknown;
  empty: boolean;
  emptyMessage?: string;
  loadingMessage?: string;
};

export function buildResourceState(input: ResourceStateInput): ResourceState {
  if (input.loading) {
    return { kind: "loading", message: input.loadingMessage ?? "Loading..." };
  }

  if (input.error) {
    return { kind: "error", message: toErrorMessage(input.error) };
  }

  if (input.empty) {
    return { kind: "empty", message: input.emptyMessage ?? "No data." };
  }

  return { kind: "ready" };
}

export function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : "请求失败";
}
