-- +goose Up
DROP INDEX IF EXISTS idx_chunks_embedding;
ALTER TABLE chunks ALTER COLUMN embedding TYPE vector;

-- +goose Down
ALTER TABLE chunks ALTER COLUMN embedding TYPE vector(1536);
CREATE INDEX idx_chunks_embedding ON chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
