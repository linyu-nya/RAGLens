package retrieval

import (
	"context"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/store"
	"github.com/pgvector/pgvector-go"
)

type Store interface {
	SearchChunks(ctx context.Context, params store.SearchChunksParams) ([]store.SearchChunksRow, error)
}

type StoreRepository struct {
	store Store
}

func NewStoreRepository(store Store) *StoreRepository {
	return &StoreRepository{store: store}
}

func (r *StoreRepository) Search(ctx context.Context, params SearchParams) ([]ChunkMatch, error) {
	rows, err := r.store.SearchChunks(ctx, store.SearchChunksParams{
		Limit:          params.TopK,
		QueryEmbedding: pgvector.NewVector(params.QueryVector),
	})
	if err != nil {
		return nil, err
	}

	items := make([]ChunkMatch, 0, len(rows))
	for _, row := range rows {
		items = append(items, ChunkMatch{
			ChunkID:       row.ID,
			DocumentID:    row.DocumentID,
			DocumentName:  row.DocumentName,
			ChunkIndex:    row.ChunkIndex,
			Content:       row.Content,
			ContentLength: row.ContentLength,
			Score:         row.Score,
		})
	}
	return items, nil
}
