export type DocumentRecord = {
  id: number;
  name: string;
  type: string;
  size: number;
  status: string;
};

export type ChunkRecord = {
  id: number;
  documentId: number;
  chunkIndex: number;
  content: string;
  contentLength: number;
};

export type ChunkMatch = {
  chunkId: number;
  documentId: number;
  documentName: string;
  chunkIndex: number;
  content: string;
  contentLength: number;
  score: number;
};

export type TraceSummary = {
  traceId: string;
  question: string;
  answer: string;
  modelName: string;
  status: string;
};

export type TraceStep = {
  stepName: string;
  stepOrder: number;
  inputData: unknown;
  outputData: unknown;
  status: string;
};

export type RetrievedTraceChunk = {
  chunkId: number;
  rank: number;
  score: number;
  contentSnapshot: string;
  documentName: string;
};

export type TraceDetail = TraceSummary & {
  finalPrompt: string;
  steps: TraceStep[];
  retrievedChunks: RetrievedTraceChunk[];
};

export type RetrievalResponse = {
  query: string;
  items: ChunkMatch[];
};

export type AskResponse = {
  traceId: string;
  question: string;
  answer: string;
  model: string;
  chunks: ChunkMatch[];
};

