import type { TraceDetail } from "./types";

export type TraceStepView = {
  order: number;
  name: string;
  status: string;
  hasInput: boolean;
  hasOutput: boolean;
  inputJson: string;
  outputJson: string;
};

export type TraceChunkView = {
  chunkId: number;
  rank: number;
  scoreLabel: string;
  documentName: string;
  content: string;
};

export type TraceViewModel = {
  heading: string;
  question: string;
  answer: string;
  promptText: string;
  hasPrompt: boolean;
  steps: TraceStepView[];
  chunks: TraceChunkView[];
};

export function buildTraceViewModel(detail: TraceDetail): TraceViewModel {
  const prompt = detail.finalPrompt.trim();

  return {
    heading: `${detail.traceId} · ${detail.status}`,
    question: detail.question,
    answer: detail.answer,
    promptText: prompt || "No final prompt recorded.",
    hasPrompt: prompt.length > 0,
    steps: [...detail.steps]
      .sort((left, right) => left.stepOrder - right.stepOrder)
      .map((step) => ({
        order: step.stepOrder,
        name: step.stepName,
        status: step.status,
        hasInput: step.inputData !== null && step.inputData !== undefined,
        hasOutput: step.outputData !== null && step.outputData !== undefined,
        inputJson: formatTracePayload(step.inputData),
        outputJson: formatTracePayload(step.outputData),
      })),
    chunks: [...detail.retrievedChunks]
      .sort((left, right) => left.rank - right.rank)
      .map((chunk) => ({
        chunkId: chunk.chunkId,
        rank: chunk.rank,
        scoreLabel: formatTraceScore(chunk.score),
        documentName: chunk.documentName,
        content: chunk.contentSnapshot,
      })),
  };
}

export function formatTraceScore(score: number) {
  return Number.isFinite(score) ? score.toFixed(4) : "n/a";
}

export function formatTracePayload(payload: unknown) {
  try {
    return JSON.stringify(payload, null, 2) ?? "null";
  } catch {
    return "Unable to render payload.";
  }
}
