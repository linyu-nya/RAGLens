package tracing

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/store"
)

const (
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type Store interface {
	CreateTrace(ctx context.Context, params store.CreateTraceParams) (store.Trace, error)
	CreateTraceStep(ctx context.Context, params store.CreateTraceStepParams) (store.TraceStep, error)
	CreateRetrievedChunk(ctx context.Context, params store.CreateRetrievedChunkParams) (store.RetrievedChunk, error)
	ListTraces(ctx context.Context, params store.ListTracesParams) ([]store.Trace, error)
	GetTraceByTraceID(ctx context.Context, traceID string) (store.Trace, error)
	ListTraceSteps(ctx context.Context, traceID string) ([]store.TraceStep, error)
	ListRetrievedChunksByTraceID(ctx context.Context, traceID string) ([]store.RetrievedChunk, error)
}

type StoreRepository struct {
	store Store
}

type RecordInput struct {
	TraceID         string
	Question        string
	FinalPrompt     string
	Answer          string
	ModelName       string
	Status          string
	ErrorMessage    string
	Steps           []StepRecord
	RetrievedChunks []RetrievedChunkRecord
}

type StepRecord struct {
	StepName     string
	StepOrder    int32
	InputData    []byte
	OutputData   []byte
	LatencyMs    int32
	Status       string
	ErrorMessage string
}

type RetrievedChunkRecord struct {
	ChunkID         int64
	Rank            int32
	Score           float64
	ContentSnapshot string
	DocumentName    string
}

type ListInput struct {
	Page     int
	PageSize int
	Status   string
	Keyword  string
}

type TraceSummary struct {
	TraceID   string
	Question  string
	Answer    string
	ModelName string
	Status    string
}

type TraceDetail struct {
	TraceID         string
	Question        string
	FinalPrompt     string
	Answer          string
	ModelName       string
	Status          string
	Steps           []TraceStepRecord
	RetrievedChunks []RetrievedChunkRecord
}

type TraceStepRecord struct {
	StepName   string
	StepOrder  int32
	InputData  []byte
	OutputData []byte
	Status     string
}

func NewStoreRepository(store Store) *StoreRepository {
	return &StoreRepository{store: store}
}

func (r *StoreRepository) Record(ctx context.Context, input RecordInput) error {
	if _, err := r.store.CreateTrace(ctx, store.CreateTraceParams{
		TraceID:        input.TraceID,
		Question:       input.Question,
		RewrittenQuery: pgtype.Text{},
		FinalPrompt:    textValue(input.FinalPrompt),
		Answer:         textValue(input.Answer),
		ModelName:      textValue(input.ModelName),
		ConfigSnapshot: []byte(`{}`),
		LatencyMs:      pgtype.Int4{},
		Status:         input.Status,
		ErrorMessage:   nullableText(input.ErrorMessage),
	}); err != nil {
		return err
	}

	for _, step := range input.Steps {
		if _, err := r.store.CreateTraceStep(ctx, store.CreateTraceStepParams{
			TraceID:      input.TraceID,
			StepName:     step.StepName,
			StepOrder:    step.StepOrder,
			InputData:    jsonData(step.InputData),
			OutputData:   jsonData(step.OutputData),
			LatencyMs:    int4Value(step.LatencyMs),
			Status:       step.Status,
			ErrorMessage: nullableText(step.ErrorMessage),
		}); err != nil {
			return err
		}
	}

	for _, chunk := range input.RetrievedChunks {
		if _, err := r.store.CreateRetrievedChunk(ctx, store.CreateRetrievedChunkParams{
			TraceID:         input.TraceID,
			ChunkID:         int8Value(chunk.ChunkID),
			Rank:            chunk.Rank,
			Score:           numericValue(chunk.Score),
			ContentSnapshot: chunk.ContentSnapshot,
			DocumentName:    chunk.DocumentName,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *StoreRepository) List(ctx context.Context, input ListInput) ([]TraceSummary, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	rows, err := r.store.ListTraces(ctx, store.ListTracesParams{
		Limit:   int32(pageSize),
		Offset:  int32((page - 1) * pageSize),
		Status:  nullableText(input.Status),
		Keyword: nullableText(input.Keyword),
	})
	if err != nil {
		return nil, err
	}

	items := make([]TraceSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, TraceSummary{
			TraceID:   row.TraceID,
			Question:  row.Question,
			Answer:    row.Answer.String,
			ModelName: row.ModelName.String,
			Status:    row.Status,
		})
	}
	return items, nil
}

func (r *StoreRepository) Get(ctx context.Context, traceID string) (TraceDetail, error) {
	trace, err := r.store.GetTraceByTraceID(ctx, traceID)
	if err != nil {
		return TraceDetail{}, err
	}
	steps, err := r.store.ListTraceSteps(ctx, traceID)
	if err != nil {
		return TraceDetail{}, err
	}
	chunks, err := r.store.ListRetrievedChunksByTraceID(ctx, traceID)
	if err != nil {
		return TraceDetail{}, err
	}

	return TraceDetail{
		TraceID:         trace.TraceID,
		Question:        trace.Question,
		FinalPrompt:     trace.FinalPrompt.String,
		Answer:          trace.Answer.String,
		ModelName:       trace.ModelName.String,
		Status:          trace.Status,
		Steps:           traceStepsFromStore(steps),
		RetrievedChunks: retrievedChunksFromStore(chunks),
	}, nil
}

func textValue(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func nullableText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return textValue(value)
}

func int4Value(value int32) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: value, Valid: true}
}

func int8Value(value int64) pgtype.Int8 {
	if value == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: value, Valid: true}
}

func numericValue(value float64) pgtype.Numeric {
	var numeric pgtype.Numeric
	_ = numeric.Scan(strconv.FormatFloat(value, 'f', -1, 64))
	return numeric
}

func float64Value(value pgtype.Numeric) float64 {
	if !value.Valid {
		return 0
	}
	result, err := value.Float64Value()
	if err != nil || !result.Valid {
		return 0
	}
	return result.Float64
}

func traceStepsFromStore(rows []store.TraceStep) []TraceStepRecord {
	items := make([]TraceStepRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, TraceStepRecord{
			StepName:   row.StepName,
			StepOrder:  row.StepOrder,
			InputData:  row.InputData,
			OutputData: row.OutputData,
			Status:     row.Status,
		})
	}
	return items
}

func retrievedChunksFromStore(rows []store.RetrievedChunk) []RetrievedChunkRecord {
	items := make([]RetrievedChunkRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, RetrievedChunkRecord{
			ChunkID:         row.ChunkID.Int64,
			Rank:            row.Rank,
			Score:           float64Value(row.Score),
			ContentSnapshot: row.ContentSnapshot,
			DocumentName:    row.DocumentName,
		})
	}
	return items
}

func jsonData(value []byte) []byte {
	if len(value) == 0 {
		return []byte(`{}`)
	}
	return value
}
