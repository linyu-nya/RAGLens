import { createContext, lazy, Suspense, useContext, useState, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createRoute, createRootRoute, createRouter, Link, Outlet, RouterProvider } from "@tanstack/react-router";
import {
  Activity,
  Braces,
  Database,
  FileText,
  Gauge,
  Home,
  Loader2,
  MessageSquareText,
  Play,
  RefreshCw,
  Search,
  Upload,
} from "lucide-react";
import { api } from "./api";
import { navigationItems } from "./navigation";
import { buildTraceViewModel } from "./traceViewModel";
import { buildResourceState, toErrorMessage, type ResourceState } from "./uiState";
import type { AskResponse, ChunkRecord, DocumentRecord, TraceDetail as TraceDetailRecord, TraceSummary } from "./types";

const defaultQuestion = "RAGLens 是什么？";
const CodeViewer = lazy(() => import("./CodeViewer").then((module) => ({ default: module.CodeViewer })));

type WorkbenchContextValue = {
  documents: DocumentRecord[];
  traces: TraceSummary[];
  chunks: ChunkRecord[];
  selectedDocument: DocumentRecord | undefined;
  selectedDocumentId: number | null;
  selectedTraceId: string | null;
  traceDetail: TraceDetailRecord | undefined;
  question: string;
  retrievalQuery: string;
  topK: number;
  lastAsk: AskResponse | null;
  retrievalItems: AskResponse["chunks"];
  notice: string;
  isBusy: boolean;
  isChunksLoading: boolean;
  isTraceDetailLoading: boolean;
  isIndexing: boolean;
  isRetrieving: boolean;
  isAsking: boolean;
  setSelectedDocumentId: (id: number) => void;
  setSelectedTraceId: (traceId: string) => void;
  setQuestion: (question: string) => void;
  setRetrievalQuery: (query: string) => void;
  setTopK: (topK: number) => void;
  uploadDocument: (file: File) => void;
  indexDocument: (documentId: number) => void;
  searchRetrieval: () => void;
  askRag: () => void;
  refreshAll: () => Promise<void>;
  health: { status?: string; app?: string };
  errors: {
    health?: unknown;
    documents?: unknown;
    chunks?: unknown;
    traces?: unknown;
    traceDetail?: unknown;
  };
  loading: {
    documents: boolean;
    traces: boolean;
  };
};

const WorkbenchContext = createContext<WorkbenchContextValue | null>(null);

const rootRoute = createRootRoute({
  component: RootLayout,
});

const overviewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: OverviewPage,
});

const documentsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "documents",
  component: DocumentsPage,
});

const askRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "ask",
  component: AskPage,
});

const tracesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "traces",
  component: TracesPage,
});

const routeTree = rootRoute.addChildren([overviewRoute, documentsRoute, askRoute, tracesRoute]);
const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export function App() {
  return <RouterProvider router={router} />;
}

function RootLayout() {
  return (
    <WorkbenchProvider>
      <AppShell>
        <Outlet />
      </AppShell>
    </WorkbenchProvider>
  );
}

function WorkbenchProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [selectedDocumentId, setSelectedDocumentIdState] = useState<number | null>(null);
  const [selectedTraceId, setSelectedTraceIdState] = useState<string | null>(null);
  const [question, setQuestion] = useState(defaultQuestion);
  const [topK, setTopK] = useState(5);
  const [retrievalQuery, setRetrievalQuery] = useState(defaultQuestion);
  const [lastAsk, setLastAsk] = useState<AskResponse | null>(null);
  const [retrievalItems, setRetrievalItems] = useState<AskResponse["chunks"]>([]);
  const [notice, setNotice] = useState<string>("");

  const healthQuery = useQuery({
    queryKey: ["health"],
    queryFn: api.health,
  });
  const documentsQuery = useQuery({
    queryKey: ["documents"],
    queryFn: api.listDocuments,
  });
  const tracesQuery = useQuery({
    queryKey: ["traces"],
    queryFn: api.listTraces,
  });
  const chunksQuery = useQuery({
    queryKey: ["chunks", selectedDocumentId],
    queryFn: () => api.listChunks(selectedDocumentId as number),
    enabled: selectedDocumentId !== null,
  });
  const traceDetailQuery = useQuery({
    queryKey: ["trace", selectedTraceId],
    queryFn: () => api.getTrace(selectedTraceId as string),
    enabled: selectedTraceId !== null,
  });

  const documents = documentsQuery.data?.items ?? [];
  const traces = tracesQuery.data?.items ?? [];
  const chunks = chunksQuery.data?.items ?? [];
  const selectedDocument = documents.find((document) => document.id === selectedDocumentId) ?? documents[0];

  const uploadMutation = useMutation({
    mutationFn: api.uploadDocument,
    onSuccess: async (document) => {
      setSelectedDocumentIdState(document.id);
      setNotice(`已上传 ${document.name}`);
      await queryClient.invalidateQueries({ queryKey: ["documents"] });
    },
    onError: (error) => setNotice(errorMessage(error)),
  });

  const indexMutation = useMutation({
    mutationFn: api.indexDocument,
    onSuccess: async (result) => {
      setNotice(`已为 ${result.chunkCount} 个 chunks 写入 embedding`);
      await queryClient.invalidateQueries({ queryKey: ["chunks", result.documentId] });
    },
    onError: (error) => setNotice(errorMessage(error)),
  });

  const retrievalMutation = useMutation({
    mutationFn: api.searchRetrieval,
    onSuccess: (result) => {
      setRetrievalItems(result.items);
      setNotice(`检索返回 ${result.items.length} 个 chunks`);
    },
    onError: (error) => setNotice(errorMessage(error)),
  });

  const askMutation = useMutation({
    mutationFn: api.askRag,
    onSuccess: async (result) => {
      setLastAsk(result);
      setSelectedTraceIdState(result.traceId);
      setNotice(`RAG Ask 完成，traceId: ${result.traceId}`);
      await queryClient.invalidateQueries({ queryKey: ["traces"] });
    },
    onError: (error) => setNotice(errorMessage(error)),
  });

  const isBusy = uploadMutation.isPending || indexMutation.isPending || retrievalMutation.isPending || askMutation.isPending;

  const value: WorkbenchContextValue = {
    documents,
    traces,
    chunks,
    selectedDocument,
    selectedDocumentId,
    selectedTraceId,
    traceDetail: traceDetailQuery.data,
    question,
    retrievalQuery,
    topK,
    lastAsk,
    retrievalItems,
    notice,
    isBusy,
    isChunksLoading: chunksQuery.isLoading,
    isTraceDetailLoading: traceDetailQuery.isLoading,
    isIndexing: indexMutation.isPending,
    isRetrieving: retrievalMutation.isPending,
    isAsking: askMutation.isPending,
    setSelectedDocumentId: setSelectedDocumentIdState,
    setSelectedTraceId: setSelectedTraceIdState,
    setQuestion,
    setRetrievalQuery,
    setTopK,
    uploadDocument: uploadMutation.mutate,
    indexDocument: indexMutation.mutate,
    searchRetrieval: () => retrievalMutation.mutate({ query: retrievalQuery, topK }),
    askRag: () => askMutation.mutate({ question, topK }),
    refreshAll: () => refreshAll(queryClient),
    health: {
      status: healthQuery.data?.status,
      app: healthQuery.data?.app,
    },
    errors: {
      health: healthQuery.error,
      documents: documentsQuery.error,
      chunks: chunksQuery.error,
      traces: tracesQuery.error,
      traceDetail: traceDetailQuery.error,
    },
    loading: {
      documents: documentsQuery.isLoading,
      traces: tracesQuery.isLoading,
    },
  };

  return <WorkbenchContext.Provider value={value}>{children}</WorkbenchContext.Provider>;
}

function useWorkbench() {
  const context = useContext(WorkbenchContext);
  if (!context) throw new Error("useWorkbench must be used within WorkbenchProvider");
  return context;
}

function AppShell({ children }: { children: ReactNode }) {
  const workbench = useWorkbench();

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">RAGLens Control Surface</p>
          <h1>RAG 调试工作台</h1>
        </div>
        <div className="status-strip" aria-label="service status">
          <span className={workbench.health.status === "ok" ? "dot is-online" : "dot"} />
          <span>{workbench.errors.health ? toErrorMessage(workbench.errors.health) : workbench.health.app ?? "RAGLens"}</span>
          <button className="icon-button" type="button" onClick={() => void workbench.refreshAll()} title="刷新">
            <RefreshCw size={16} />
          </button>
        </div>
      </header>

      <nav className="app-nav" aria-label="Primary">
        {navigationItems.map((item) => (
          <Link key={item.to} to={item.to} className="nav-link" activeProps={{ className: "nav-link is-active" }}>
            {item.label}
          </Link>
        ))}
      </nav>

      {workbench.notice ? <div className="notice">{workbench.notice}</div> : null}

      {children}

      {workbench.isBusy ? (
        <div className="work-indicator">
          <Loader2 className="spin" size={15} />
          running request
        </div>
      ) : null}
    </main>
  );
}

function OverviewPage() {
  const workbench = useWorkbench();

  return (
    <section className="overview-grid">
      <OverviewCard icon={<Database size={17} />} label="Documents" value={`${workbench.documents.length}`} to="/documents" text="上传语料、查看 chunks、写入 embedding。" />
      <OverviewCard icon={<MessageSquareText size={17} />} label="Ask" value={workbench.lastAsk?.traceId ?? "ready"} to="/ask" text="检索调试和最小 RAG 问答入口。" />
      <OverviewCard icon={<Activity size={17} />} label="Traces" value={`${workbench.traces.length}`} to="/traces" text="追踪 prompt、steps 和 retrieved chunks。" />
      <section className="panel overview-wide">
        <PanelTitle icon={<Home size={17} />} title="Current Run" metric={workbench.health.status ?? "offline"} />
        <div className="overview-copy">
          <strong>{workbench.lastAsk?.question ?? "还没有 RAG Ask 结果"}</strong>
          <p>{workbench.lastAsk?.answer ?? "先进入 Ask 页面提问；完成后这里会显示最近一次回答和 traceId。"}</p>
        </div>
      </section>
    </section>
  );
}

function OverviewCard({ icon, label, value, to, text }: { icon: ReactNode; label: string; value: string; to: "/" | "/documents" | "/ask" | "/traces"; text: string }) {
  return (
    <Link to={to} className="panel overview-card">
      <PanelTitle icon={icon} title={label} metric={value} />
      <p>{text}</p>
    </Link>
  );
}

function DocumentsPage() {
  const workbench = useWorkbench();

  return (
    <section className="page-grid documents-layout">
      <aside className="panel documents-panel">
        <PanelTitle icon={<Database size={17} />} title="Documents" metric={`${workbench.documents.length} files`} />
        <label className="upload-zone">
          <Upload size={18} />
          <span>上传 .md / .txt</span>
          <input
            type="file"
            accept=".md,.txt,text/markdown,text/plain"
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (file) workbench.uploadDocument(file);
              event.currentTarget.value = "";
            }}
          />
        </label>
        <DocumentList
          documents={workbench.documents}
          selectedId={workbench.selectedDocumentId ?? workbench.selectedDocument?.id ?? null}
          onSelect={workbench.setSelectedDocumentId}
          loading={workbench.loading.documents}
          error={workbench.errors.documents}
        />
      </aside>

      <section className="panel document-detail">
        <div className="section-head">
          <PanelTitle icon={<FileText size={17} />} title={workbench.selectedDocument?.name ?? "No document selected"} metric={workbench.selectedDocument?.status ?? "idle"} />
          <button
            className="command-button"
            type="button"
            disabled={!workbench.selectedDocument || workbench.isIndexing}
            onClick={() => workbench.selectedDocument && workbench.indexDocument(workbench.selectedDocument.id)}
          >
            {workbench.isIndexing ? <Loader2 className="spin" size={16} /> : <Play size={16} />}
            Index
          </button>
        </div>
        <ChunkPreview chunks={workbench.chunks} loading={workbench.isChunksLoading} error={workbench.errors.chunks} hasSelectedDocument={Boolean(workbench.selectedDocument)} />
      </section>
    </section>
  );
}

function AskPage() {
  const workbench = useWorkbench();

  return (
    <section className="page-grid ask-layout">
      <section className="panel">
        <PanelTitle icon={<Search size={17} />} title="Retrieval Search" metric={`topK ${workbench.topK}`} />
        <textarea className="question-box compact" value={workbench.retrievalQuery} onChange={(event) => workbench.setRetrievalQuery(event.target.value)} />
        <ActionRow topK={workbench.topK} setTopK={workbench.setTopK} disabled={workbench.isRetrieving} onSubmit={workbench.searchRetrieval} label="Search" />
        <MatchList items={workbench.retrievalItems} />
      </section>

      <section className="panel ask-panel">
        <PanelTitle icon={<MessageSquareText size={17} />} title="RAG Ask" metric={workbench.lastAsk?.traceId ?? "no trace"} />
        <textarea className="question-box" value={workbench.question} onChange={(event) => workbench.setQuestion(event.target.value)} />
        <ActionRow topK={workbench.topK} setTopK={workbench.setTopK} disabled={workbench.isAsking} onSubmit={workbench.askRag} label="Ask" />
        <AnswerBlock answer={workbench.lastAsk} pending={workbench.isAsking} />
      </section>
    </section>
  );
}

function TracesPage() {
  const workbench = useWorkbench();

  return (
    <section className="page-grid traces-layout">
      <aside className="panel traces-panel">
        <PanelTitle icon={<Activity size={17} />} title="Traces" metric={`${workbench.traces.length} recent`} />
        <TraceList traces={workbench.traces} selectedId={workbench.selectedTraceId} onSelect={workbench.setSelectedTraceId} loading={workbench.loading.traces} error={workbench.errors.traces} />
      </aside>
      <section className="panel trace-reader">
        <PanelTitle icon={<Braces size={17} />} title="Trace Detail" metric={workbench.selectedTraceId ?? "select trace"} />
        <TraceDetail detail={workbench.traceDetail} loading={workbench.isTraceDetailLoading} error={workbench.errors.traceDetail} selected={Boolean(workbench.selectedTraceId)} />
      </section>
    </section>
  );
}

function PanelTitle({ icon, title, metric }: { icon: ReactNode; title: string; metric: string }) {
  return (
    <div className="panel-title">
      <span className="title-icon">{icon}</span>
      <span>{title}</span>
      <em>{metric}</em>
    </div>
  );
}

function DocumentList({
  documents,
  selectedId,
  onSelect,
  loading,
  error,
}: {
  documents: DocumentRecord[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  loading: boolean;
  error: unknown;
}) {
  const state = buildResourceState({ loading, error, empty: documents.length === 0, emptyMessage: "还没有文档。上传 .md 或 .txt 后开始调试。" });
  if (state.kind !== "ready") return <ResourceStateBlock state={state} />;

  return (
    <div className="document-list">
      {documents.map((document) => (
        <button key={document.id} className={document.id === selectedId ? "list-item is-selected" : "list-item"} type="button" onClick={() => onSelect(document.id)}>
          <strong>{document.name}</strong>
          <span>{document.type} · {formatBytes(document.size)} · {document.status}</span>
        </button>
      ))}
    </div>
  );
}

function ChunkPreview({ chunks, loading, error, hasSelectedDocument }: { chunks: ChunkRecord[]; loading: boolean; error: unknown; hasSelectedDocument: boolean }) {
  const state = buildResourceState({
    loading,
    error,
    empty: !hasSelectedDocument || chunks.length === 0,
    emptyMessage: hasSelectedDocument ? "这个文档还没有 chunks，确认上传内容是否可解析。" : "选择文档后查看 chunks。",
    loadingMessage: "加载 chunks...",
  });
  if (state.kind !== "ready") return <ResourceStateBlock state={state} />;

  return (
    <div className="chunk-grid">
      {chunks.slice(0, 9).map((chunk) => (
        <article key={chunk.id} className="chunk-card">
          <span>#{chunk.chunkIndex}</span>
          <p>{chunk.content}</p>
        </article>
      ))}
    </div>
  );
}

function ActionRow({ topK, setTopK, disabled, onSubmit, label }: { topK: number; setTopK: (value: number) => void; disabled: boolean; onSubmit: () => void; label: string }) {
  return (
    <div className="action-row">
      <label>
        <Gauge size={15} />
        <input type="number" min={1} max={20} value={topK} onChange={(event) => setTopK(Number(event.target.value))} />
      </label>
      <button className="command-button" type="button" disabled={disabled} onClick={onSubmit}>
        {disabled ? <Loader2 className="spin" size={16} /> : <Play size={16} />}
        {label}
      </button>
    </div>
  );
}

function MatchList({ items }: { items: AskResponse["chunks"] }) {
  if (items.length === 0) return <p className="empty">还没有检索结果。</p>;
  return (
    <div className="match-list">
      {items.map((item) => (
        <article key={`${item.documentName}-${item.chunkId}`} className="match-card">
          <span>{item.documentName} #{item.chunkIndex}</span>
          <strong>{item.score.toFixed(4)}</strong>
          <p>{item.content}</p>
        </article>
      ))}
    </div>
  );
}

function AnswerBlock({ answer, pending }: { answer: AskResponse | null; pending: boolean }) {
  if (pending) return <p className="empty">正在生成回答...</p>;
  if (!answer) return <p className="empty">提出问题后查看答案、引用 chunks 和 traceId。</p>;
  return (
    <article className="answer-block">
      <span>{answer.model} · {answer.traceId}</span>
      <p>{answer.answer}</p>
    </article>
  );
}

function TraceList({
  traces,
  selectedId,
  onSelect,
  loading,
  error,
}: {
  traces: TraceSummary[];
  selectedId: string | null;
  onSelect: (traceId: string) => void;
  loading: boolean;
  error: unknown;
}) {
  const state = buildResourceState({ loading, error, empty: traces.length === 0, emptyMessage: "还没有 trace。先在 Ask 页面发起一次 RAG Ask。" });
  if (state.kind !== "ready") return <ResourceStateBlock state={state} />;

  return (
    <div className="trace-list">
      {traces.map((trace) => (
        <button key={trace.traceId} className={trace.traceId === selectedId ? "list-item is-selected" : "list-item"} type="button" onClick={() => onSelect(trace.traceId)}>
          <strong>{trace.question}</strong>
          <span>{trace.status} · {trace.modelName || "model pending"}</span>
        </button>
      ))}
    </div>
  );
}

function TraceDetail({ detail, loading, error, selected }: { detail: TraceDetailRecord | undefined; loading: boolean; error: unknown; selected: boolean }) {
  const state = buildResourceState({
    loading,
    error,
    empty: !selected || !detail,
    emptyMessage: selected ? "Trace 详情为空。" : "选择 trace 查看 prompt、steps 和 retrieved chunks。",
    loadingMessage: "加载 trace...",
  });
  if (state.kind !== "ready") return <ResourceStateBlock state={state} />;
  if (!detail) return null;

  const trace = buildTraceViewModel(detail);

  return (
    <div className="trace-detail">
      <div className="trace-summary">
        <span>{trace.heading}</span>
        <strong>{trace.question}</strong>
        <p>{trace.answer || "No answer recorded."}</p>
      </div>
      <div className="code-label"><Braces size={14} /> final prompt</div>
      <Suspense fallback={<div className="code-viewer code-viewer-fallback">Loading viewer...</div>}>
        <CodeViewer value={trace.promptText} language="text" className={trace.hasPrompt ? "" : "is-empty"} />
      </Suspense>
      <div className="step-stack">
        {trace.steps.map((step) => (
          <div key={`${step.order}-${step.name}`} className="step-row">
            <span>{step.order}</span>
            <strong>{step.name}</strong>
            <em>{step.status}</em>
            <small>{step.hasInput ? "input" : "no input"} · {step.hasOutput ? "output" : "no output"}</small>
            <details className="step-payload">
              <summary>payload</summary>
              <div className="payload-grid">
                <section>
                  <div className="code-label">input</div>
                  <Suspense fallback={<div className="code-viewer code-viewer-fallback">Loading viewer...</div>}>
                    <CodeViewer value={step.inputJson} language="json" />
                  </Suspense>
                </section>
                <section>
                  <div className="code-label">output</div>
                  <Suspense fallback={<div className="code-viewer code-viewer-fallback">Loading viewer...</div>}>
                    <CodeViewer value={step.outputJson} language="json" />
                  </Suspense>
                </section>
              </div>
            </details>
          </div>
        ))}
      </div>
      <div className="code-label"><Search size={14} /> retrieved chunks</div>
      {trace.chunks.length === 0 ? (
        <p className="empty compact-empty">No retrieved chunks recorded.</p>
      ) : (
        <div className="trace-chunks">
          {trace.chunks.map((chunk) => (
            <article key={`${chunk.rank}-${chunk.chunkId}`} className="trace-chunk">
              <span>#{chunk.rank} · {chunk.documentName}</span>
              <strong>{chunk.scoreLabel}</strong>
              <p>{chunk.content}</p>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

function ResourceStateBlock({ state }: { state: Exclude<ResourceState, { kind: "ready" }> }) {
  return (
    <div className={`resource-state is-${state.kind}`}>
      {state.kind === "loading" ? <Loader2 className="spin" size={16} /> : null}
      <span>{state.message}</span>
    </div>
  );
}

function formatBytes(size: number) {
  if (size < 1024) return `${size} B`;
  return `${(size / 1024).toFixed(1)} KiB`;
}

function errorMessage(error: unknown) {
  return toErrorMessage(error);
}

async function refreshAll(queryClient: ReturnType<typeof useQueryClient>) {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ["health"] }),
    queryClient.invalidateQueries({ queryKey: ["documents"] }),
    queryClient.invalidateQueries({ queryKey: ["traces"] }),
  ]);
}
