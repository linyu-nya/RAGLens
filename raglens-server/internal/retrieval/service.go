package retrieval

import (
	"context"
	"errors"
	"strings"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/embedding"
)

const (
	DefaultTopK = 5
	MaxTopK     = 20
)

var (
	ErrEmptyQuery                 = errors.New("empty query")
	ErrEmbeddingCountMismatch     = errors.New("embedding count mismatch")
	ErrEmbeddingDimensionMismatch = errors.New("embedding dimension mismatch")
)

type Repository interface {
	Search(ctx context.Context, params SearchParams) ([]ChunkMatch, error)
}

type EmbeddingClient interface {
	Embed(ctx context.Context, input embedding.Input) ([]embedding.Vector, error)
}

type Options struct {
	EmbeddingDimensions int
}

type Service struct {
	repository          Repository
	embedder            EmbeddingClient
	embeddingDimensions int
}

type SearchInput struct {
	Query string
	TopK  int
}

type SearchParams struct {
	QueryVector []float32
	TopK        int32
}

type SearchResult struct {
	Query string
	Items []ChunkMatch
}

type ChunkMatch struct {
	ChunkID       int64
	DocumentID    int64
	DocumentName  string
	ChunkIndex    int32
	Content       string
	ContentLength int32
	Score         float64
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

func (s *Service) Search(ctx context.Context, input SearchInput) (SearchResult, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return SearchResult{}, ErrEmptyQuery
	}

	vectors, err := s.embedder.Embed(ctx, embedding.Input{Texts: []string{query}})
	if err != nil {
		return SearchResult{}, err
	}
	if len(vectors) != 1 {
		return SearchResult{}, ErrEmbeddingCountMismatch
	}
	if s.embeddingDimensions > 0 && len(vectors[0]) != s.embeddingDimensions {
		return SearchResult{}, ErrEmbeddingDimensionMismatch
	}

	topK := normalizeTopK(input.TopK)
	items, err := s.repository.Search(ctx, SearchParams{
		QueryVector: []float32(vectors[0]),
		TopK:        int32(topK),
	})
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{
		Query: query,
		Items: items,
	}, nil
}

func normalizeTopK(value int) int {
	if value < 1 {
		return DefaultTopK
	}
	if value > MaxTopK {
		return MaxTopK
	}
	return value
}
