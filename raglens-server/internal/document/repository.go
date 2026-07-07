package document

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/store"
)

type Store interface {
	CreateChunk(ctx context.Context, params store.CreateChunkParams) (store.Chunk, error)
	CreateDocument(ctx context.Context, params store.CreateDocumentParams) (store.Document, error)
	ListChunksByDocument(ctx context.Context, documentID int64) ([]store.Chunk, error)
	ListDocuments(ctx context.Context, params store.ListDocumentsParams) ([]store.Document, error)
}

type StoreRepository struct {
	store Store
}

func NewStoreRepository(store Store) *StoreRepository {
	return &StoreRepository{store: store}
}

func (r *StoreRepository) Create(ctx context.Context, input CreateRecord) (Record, error) {
	row, err := r.store.CreateDocument(ctx, store.CreateDocumentParams{
		Name:         input.Name,
		Type:         input.Type,
		Size:         input.Size,
		Status:       input.Status,
		RawText:      textValue(input.RawText),
		ErrorMessage: pgtype.Text{},
	})
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:      row.ID,
		Name:    row.Name,
		Type:    row.Type,
		Size:    row.Size,
		Status:  row.Status,
		RawText: row.RawText.String,
	}, nil
}

func (r *StoreRepository) List(ctx context.Context, params ListParams) ([]Record, error) {
	rows, err := r.store.ListDocuments(ctx, store.ListDocumentsParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, err
	}

	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, recordFromStore(row))
	}
	return records, nil
}

func (r *StoreRepository) CreateChunks(ctx context.Context, documentID int64, chunks []CreateChunkRecord) error {
	for _, chunk := range chunks {
		if _, err := r.store.CreateChunk(ctx, store.CreateChunkParams{
			DocumentID:    documentID,
			ChunkIndex:    chunk.ChunkIndex,
			Content:       chunk.Content,
			ContentLength: chunk.ContentLength,
		}); err != nil {
			return err
		}
	}
	return nil
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

func recordFromStore(row store.Document) Record {
	return Record{
		ID:      row.ID,
		Name:    row.Name,
		Type:    row.Type,
		Size:    row.Size,
		Status:  row.Status,
		RawText: row.RawText.String,
	}
}

func textValue(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}
