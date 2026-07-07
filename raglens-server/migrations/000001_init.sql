-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    size BIGINT NOT NULL,
    status TEXT NOT NULL,
    raw_text TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chunks (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    content_length INTEGER NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (document_id, chunk_index)
);

CREATE TABLE traces (
    id BIGSERIAL PRIMARY KEY,
    trace_id TEXT NOT NULL UNIQUE,
    question TEXT NOT NULL,
    rewritten_query TEXT,
    final_prompt TEXT,
    answer TEXT,
    model_name TEXT,
    config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    latency_ms INTEGER,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    status TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE trace_steps (
    id BIGSERIAL PRIMARY KEY,
    trace_id TEXT NOT NULL REFERENCES traces(trace_id) ON DELETE CASCADE,
    step_name TEXT NOT NULL,
    step_order INTEGER NOT NULL,
    input_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    output_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    latency_ms INTEGER,
    status TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (trace_id, step_order)
);

CREATE TABLE retrieved_chunks (
    id BIGSERIAL PRIMARY KEY,
    trace_id TEXT NOT NULL REFERENCES traces(trace_id) ON DELETE CASCADE,
    chunk_id BIGINT REFERENCES chunks(id) ON DELETE SET NULL,
    rank INTEGER NOT NULL,
    score NUMERIC,
    content_snapshot TEXT NOT NULL,
    document_name TEXT NOT NULL
);

CREATE TABLE eval_reports (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    total_cases INTEGER NOT NULL,
    passed_cases INTEGER NOT NULL,
    failed_cases INTEGER NOT NULL,
    keyword_hit_rate NUMERIC,
    avg_latency_ms INTEGER,
    avg_tokens INTEGER,
    report_markdown TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_chunks_document_id ON chunks(document_id);
CREATE INDEX idx_chunks_embedding ON chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_traces_created_at ON traces(created_at DESC);
CREATE INDEX idx_traces_status ON traces(status);
CREATE INDEX idx_trace_steps_trace_id ON trace_steps(trace_id);
CREATE INDEX idx_retrieved_chunks_trace_id ON retrieved_chunks(trace_id);

-- +goose Down
DROP TABLE IF EXISTS eval_reports;
DROP TABLE IF EXISTS retrieved_chunks;
DROP TABLE IF EXISTS trace_steps;
DROP TABLE IF EXISTS traces;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS documents;
DROP EXTENSION IF EXISTS vector;
