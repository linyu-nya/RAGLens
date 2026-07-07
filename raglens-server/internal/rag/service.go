package rag

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/chat"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/retrieval"
)

var ErrEmptyQuestion = errors.New("empty question")

type Retriever interface {
	Search(ctx context.Context, input retrieval.SearchInput) (retrieval.SearchResult, error)
}

type ChatClient interface {
	Complete(ctx context.Context, input chat.Input) (chat.Result, error)
}

type Recorder interface {
	Record(ctx context.Context, input TraceRecord) error
}

type Service struct {
	retriever        Retriever
	chat             ChatClient
	recorder         Recorder
	traceIDGenerator func() string
}

type Options struct {
	Recorder         Recorder
	TraceIDGenerator func() string
}

type AskInput struct {
	Question string
	TopK     int
}

type AskResult struct {
	TraceID  string
	Question string
	Answer   string
	Model    string
	Chunks   []retrieval.ChunkMatch
}

type TraceRecord struct {
	TraceID     string
	Question    string
	FinalPrompt string
	Answer      string
	ModelName   string
	Chunks      []retrieval.ChunkMatch
}

func NewService(retriever Retriever, chatClient ChatClient) *Service {
	return NewServiceWithOptions(retriever, chatClient, Options{})
}

func NewServiceWithOptions(retriever Retriever, chatClient ChatClient, options Options) *Service {
	traceIDGenerator := options.TraceIDGenerator
	if traceIDGenerator == nil {
		traceIDGenerator = newTraceID
	}

	return &Service{
		retriever:        retriever,
		chat:             chatClient,
		recorder:         options.Recorder,
		traceIDGenerator: traceIDGenerator,
	}
}

func (s *Service) Ask(ctx context.Context, input AskInput) (AskResult, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		return AskResult{}, ErrEmptyQuestion
	}

	retrieved, err := s.retriever.Search(ctx, retrieval.SearchInput{
		Query: question,
		TopK:  input.TopK,
	})
	if err != nil {
		return AskResult{}, err
	}

	finalPrompt := userPrompt(question, retrieved.Items)
	answer, err := s.chat.Complete(ctx, chat.Input{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt()},
			{Role: chat.RoleUser, Content: finalPrompt},
		},
	})
	if err != nil {
		return AskResult{}, err
	}

	traceID := s.traceIDGenerator()
	if s.recorder != nil {
		if err := s.recorder.Record(ctx, TraceRecord{
			TraceID:     traceID,
			Question:    question,
			FinalPrompt: finalPrompt,
			Answer:      answer.Content,
			ModelName:   answer.Model,
			Chunks:      retrieved.Items,
		}); err != nil {
			return AskResult{}, err
		}
	}

	return AskResult{
		TraceID:  traceID,
		Question: question,
		Answer:   answer.Content,
		Model:    answer.Model,
		Chunks:   retrieved.Items,
	}, nil
}

func systemPrompt() string {
	return "You are RAGLens, a careful RAG assistant. Answer using only the provided context. If the context is insufficient, say you do not know."
}

func userPrompt(question string, chunks []retrieval.ChunkMatch) string {
	var builder strings.Builder
	builder.WriteString("Question: ")
	builder.WriteString(question)
	builder.WriteString("\n\nContext:\n")

	for index, chunk := range chunks {
		_, _ = fmt.Fprintf(&builder, "[%d] %s#%d score=%.4f\n%s\n\n", index+1, chunk.DocumentName, chunk.ChunkIndex, chunk.Score, chunk.Content)
	}

	builder.WriteString("Answer in Chinese when the user asks in Chinese. Cite relevant context numbers when useful.")
	return builder.String()
}

func newTraceID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "trace-unknown"
	}
	return "trace-" + hex.EncodeToString(bytes[:])
}

func TraceRecordJSON(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return data
}
