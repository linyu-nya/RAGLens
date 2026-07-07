package indexing

import (
	"context"
	"errors"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/embedding"
)

var (
	ErrInvalidDocumentID          = errors.New("invalid document id")
	ErrEmbeddingCountMismatch     = errors.New("embedding count mismatch")
	ErrEmbeddingDimensionMismatch = errors.New("embedding dimension mismatch")
)

type Repository interface {
	ListChunks(ctx context.Context, documentID int64) ([]ChunkRecord, error)
	UpdateChunkEmbedding(ctx context.Context, chunkID int64, vector []float32) error
}

type EmbeddingClient interface {
	Embed(ctx context.Context, input embedding.Input) ([]embedding.Vector, error)
}

type Service struct {
	repository          Repository
	embedder            EmbeddingClient
	embeddingDimensions int
}

type Options struct {
	EmbeddingDimensions int
}

type IndexInput struct {
	DocumentID int64
}

type IndexResult struct {
	DocumentID int64
	ChunkCount int
}

func NewService(repository Repository, embedder EmbeddingClient) *Service {
	return NewServiceWithOptions(repository, embedder, Options{})
}

func NewServiceWithOptions(repository Repository, embedder EmbeddingClient, options Options) *Service {
	return &Service{
		repository:          repository,
		embedder:            embedder,
		embeddingDimensions: options.EmbeddingDimensions,
	}
}

func (s *Service) IndexDocument(ctx context.Context, input IndexInput) (IndexResult, error) {
	if input.DocumentID <= 0 {
		return IndexResult{}, ErrInvalidDocumentID
	}

	chunks, err := s.repository.ListChunks(ctx, input.DocumentID)
	if err != nil {
		return IndexResult{}, err
	}
	if len(chunks) == 0 {
		return IndexResult{DocumentID: input.DocumentID}, nil
	}

	texts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		texts = append(texts, chunk.Content)
	}

	vectors, err := s.embedder.Embed(ctx, embedding.Input{Texts: texts})
	if err != nil {
		return IndexResult{}, err
	}
	if len(vectors) != len(chunks) {
		return IndexResult{}, ErrEmbeddingCountMismatch
	}

	for index, vector := range vectors {
		if s.embeddingDimensions > 0 && len(vector) != s.embeddingDimensions {
			return IndexResult{}, ErrEmbeddingDimensionMismatch
		}
		if err := s.repository.UpdateChunkEmbedding(ctx, chunks[index].ID, []float32(vector)); err != nil {
			return IndexResult{}, err
		}
	}

	return IndexResult{
		DocumentID: input.DocumentID,
		ChunkCount: len(chunks),
	}, nil
}
