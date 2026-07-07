package indexing

import (
	"context"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/store"
	"github.com/pgvector/pgvector-go"
)

type Store interface {
	ListChunksByDocument(ctx context.Context, documentID int64) ([]store.Chunk, error)
	UpdateChunkEmbedding(ctx context.Context, params store.UpdateChunkEmbeddingParams) error
}

type StoreRepository struct {
	store Store
}

type ChunkRecord struct {
	ID            int64
	DocumentID    int64
	ChunkIndex    int32
	Content       string
	ContentLength int32
}

func NewStoreRepository(store Store) *StoreRepository {
	return &StoreRepository{store: store}
}

func (r *StoreRepository) ListChunks(ctx context.Context, documentID int64) ([]ChunkRecord, error) {
	rows, err := r.store.ListChunksByDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}

	chunks := make([]ChunkRecord, 0, len(rows))
	for _, row := range rows {
		chunks = append(chunks, ChunkRecord{
			ID:            row.ID,
			DocumentID:    row.DocumentID,
			ChunkIndex:    row.ChunkIndex,
			Content:       row.Content,
			ContentLength: row.ContentLength,
		})
	}
	return chunks, nil
}

func (r *StoreRepository) UpdateChunkEmbedding(ctx context.Context, chunkID int64, vector []float32) error {
	return r.store.UpdateChunkEmbedding(ctx, store.UpdateChunkEmbeddingParams{
		ID:        chunkID,
		Embedding: pgvector.NewVector(vector),
	})
}

