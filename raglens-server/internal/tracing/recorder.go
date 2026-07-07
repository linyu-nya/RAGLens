package tracing

import (
	"context"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/rag"
)

type Repository interface {
	Record(ctx context.Context, input RecordInput) error
}

type Recorder struct {
	repository Repository
}

func NewRecorder(repository Repository) *Recorder {
	return &Recorder{repository: repository}
}

func (r *Recorder) Record(ctx context.Context, input rag.TraceRecord) error {
	chunks := make([]RetrievedChunkRecord, 0, len(input.Chunks))
	for index, chunk := range input.Chunks {
		chunks = append(chunks, RetrievedChunkRecord{
			ChunkID:         chunk.ChunkID,
			Rank:            int32(index + 1),
			Score:           chunk.Score,
			ContentSnapshot: chunk.Content,
			DocumentName:    chunk.DocumentName,
		})
	}

	return r.repository.Record(ctx, RecordInput{
		TraceID:     input.TraceID,
		Question:    input.Question,
		FinalPrompt: input.FinalPrompt,
		Answer:      input.Answer,
		ModelName:   input.ModelName,
		Status:      StatusSucceeded,
		Steps: []StepRecord{
			{
				StepName:  "retrieval",
				StepOrder: 1,
				OutputData: rag.TraceRecordJSON(map[string]any{
					"chunkCount": len(input.Chunks),
				}),
				Status: StatusSucceeded,
			},
			{
				StepName:  "chat",
				StepOrder: 2,
				InputData: rag.TraceRecordJSON(map[string]any{
					"prompt": input.FinalPrompt,
				}),
				OutputData: rag.TraceRecordJSON(map[string]any{
					"answer": input.Answer,
					"model":  input.ModelName,
				}),
				Status: StatusSucceeded,
			},
		},
		RetrievedChunks: chunks,
	})
}
