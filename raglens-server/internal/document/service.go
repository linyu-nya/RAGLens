package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	chunksplit "github.com/linyu-nya/RAGLens/raglens-server/internal/chunk"
)

const (
	StatusUploaded = "uploaded"
	StatusParsed   = "parsed"

	DefaultMaxUploadBytes int64 = 10 * 1024 * 1024
	DefaultPageSize             = 20
	MaxPageSize                 = 100
)

var (
	ErrUnsupportedType   = errors.New("unsupported document type")
	ErrFileTooLarge      = errors.New("file too large")
	ErrInvalidDocumentID = errors.New("invalid document id")
)

type Repository interface {
	Create(ctx context.Context, input CreateRecord) (Record, error)
	CreateChunks(ctx context.Context, documentID int64, chunks []CreateChunkRecord) error
	List(ctx context.Context, params ListParams) ([]Record, error)
	ListChunks(ctx context.Context, documentID int64) ([]ChunkRecord, error)
}

type Service struct {
	repository     Repository
	maxUploadBytes int64
	splitter       *chunksplit.Splitter
}

type UploadInput struct {
	Filename string
	Size     int64
	Content  io.Reader
}

type ListInput struct {
	Page     int
	PageSize int
}

type ListParams struct {
	Limit  int32
	Offset int32
}

type CreateRecord struct {
	Name    string
	Type    string
	Size    int64
	Status  string
	RawText string
}

type CreateChunkRecord struct {
	DocumentID    int64
	ChunkIndex    int32
	Content       string
	ContentLength int32
}

type ChunkRecord struct {
	ID            int64
	DocumentID    int64
	ChunkIndex    int32
	Content       string
	ContentLength int32
}

type Record struct {
	ID      int64
	Name    string
	Type    string
	Size    int64
	Status  string
	RawText string
}

func NewService(repository Repository) *Service {
	return &Service{
		repository:     repository,
		maxUploadBytes: DefaultMaxUploadBytes,
		splitter:       chunksplit.NewSplitter(chunksplit.Options{}),
	}
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (Record, error) {
	documentType, err := supportedType(input.Filename)
	if err != nil {
		return Record{}, err
	}
	if input.Size > s.maxUploadBytes {
		return Record{}, ErrFileTooLarge
	}

	limited := io.LimitReader(input.Content, s.maxUploadBytes+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return Record{}, fmt.Errorf("read document content: %w", err)
	}
	if int64(len(content)) > s.maxUploadBytes {
		return Record{}, ErrFileTooLarge
	}

	record, err := s.repository.Create(ctx, CreateRecord{
		Name:    input.Filename,
		Type:    documentType,
		Size:    input.Size,
		Status:  StatusParsed,
		RawText: string(content),
	})
	if err != nil {
		return Record{}, err
	}

	chunks, err := s.splitter.Split(string(content))
	if err != nil {
		return Record{}, err
	}
	if len(chunks) == 0 {
		return record, nil
	}

	chunkRecords := make([]CreateChunkRecord, 0, len(chunks))
	for _, chunk := range chunks {
		chunkRecords = append(chunkRecords, CreateChunkRecord{
			DocumentID:    record.ID,
			ChunkIndex:    int32(chunk.Index),
			Content:       chunk.Content,
			ContentLength: int32(chunk.ContentLength),
		})
	}
	if err := s.repository.CreateChunks(ctx, record.ID, chunkRecords); err != nil {
		return Record{}, err
	}

	return record, nil
}

func (s *Service) List(ctx context.Context, input ListInput) ([]Record, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return s.repository.List(ctx, ListParams{
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
}

func (s *Service) ListChunks(ctx context.Context, documentID int64) ([]ChunkRecord, error) {
	if documentID <= 0 {
		return nil, ErrInvalidDocumentID
	}
	return s.repository.ListChunks(ctx, documentID)
}

func supportedType(filename string) (string, error) {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch extension {
	case "txt", "md":
		return extension, nil
	default:
		return "", ErrUnsupportedType
	}
}
