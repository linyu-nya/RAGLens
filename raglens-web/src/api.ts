import type {
  AskResponse,
  ChunkRecord,
  DocumentRecord,
  RetrievalResponse,
  TraceDetail,
  TraceSummary,
} from "./types";

type Fetcher = typeof fetch;

export type ApiClientOptions = {
  baseUrl?: string;
  fetcher?: Fetcher;
};

type ListResponse<T> = {
  items: T[];
};

type UploadResponse = DocumentRecord;

type IndexResponse = {
  documentId: number;
  chunkCount: number;
};

export type SearchInput = {
  query: string;
  topK: number;
};

export type AskInput = {
  question: string;
  topK: number;
};

export function createApiClient(options: ApiClientOptions = {}) {
  const baseUrl = options.baseUrl ?? "";
  const fetcher = options.fetcher ?? fetch;

  async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetcher(`${baseUrl}${path}`, init);
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const message = typeof data?.error === "string" ? data.error : `Request failed with status ${response.status}`;
      throw new Error(message);
    }
    return data as T;
  }

  return {
    health: () => request<{ status: string; app: string }>("/healthz"),
    uploadDocument: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<UploadResponse>("/api/documents/upload", {
        method: "POST",
        body: form,
      });
    },
    listDocuments: () => request<ListResponse<DocumentRecord>>("/api/documents?page=1&pageSize=50"),
    listChunks: (documentId: number) => request<ListResponse<ChunkRecord>>(`/api/documents/${documentId}/chunks`),
    indexDocument: (documentId: number) =>
      request<IndexResponse>(`/api/documents/${documentId}/index`, {
        method: "POST",
      }),
    searchRetrieval: (input: SearchInput) =>
      request<RetrievalResponse>("/api/retrieval/search", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
      }),
    askRag: (input: AskInput) =>
      request<AskResponse>("/api/rag/ask", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
      }),
    listTraces: () => request<ListResponse<TraceSummary>>("/api/traces?page=1&pageSize=20"),
    getTrace: (traceId: string) => request<TraceDetail>(`/api/traces/${traceId}`),
  };
}

export const api = createApiClient();

