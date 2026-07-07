package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/chat"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/config"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/database"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/document"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/embedding"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/httpapi"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/indexing"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/rag"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/retrieval"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/store"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/tracing"
)

func main() {
	cfg := config.Load()
	poolConfig, err := database.NewPoolConfig(cfg.Database)
	if err != nil {
		log.Fatalf("database config error: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer pool.Close()

	queries := store.New(pool)
	documentRepository := document.NewStoreRepository(queries)
	documentService := document.NewService(documentRepository)
	embeddingClient := embedding.NewOpenAIClient(embedding.OpenAIOptions{
		BaseURL: cfg.Models.OpenAIBaseURL,
		APIKey:  cfg.Models.OpenAIAPIKey,
		Model:   cfg.Models.EmbeddingModel,
	})
	indexingRepository := indexing.NewStoreRepository(queries)
	indexingService := indexing.NewServiceWithOptions(indexingRepository, embeddingClient, indexing.Options{
		EmbeddingDimensions: cfg.Models.EmbeddingDimensions,
	})
	retrievalRepository := retrieval.NewStoreRepository(queries)
	retrievalService := retrieval.NewServiceWithOptions(retrievalRepository, embeddingClient, retrieval.Options{
		EmbeddingDimensions: cfg.Models.EmbeddingDimensions,
	})
	chatClient := chat.NewOpenAIClient(chat.OpenAIOptions{
		BaseURL: cfg.Models.OpenAIBaseURL,
		APIKey:  cfg.Models.OpenAIAPIKey,
		Model:   cfg.Models.ChatModel,
	})
	tracingRepository := tracing.NewStoreRepository(queries)
	tracingRecorder := tracing.NewRecorder(tracingRepository)
	ragService := rag.NewServiceWithOptions(retrievalService, chatClient, rag.Options{
		Recorder: tracingRecorder,
	})

	server := &http.Server{
		Addr: cfg.Server.Addr,
		Handler: httpapi.NewRouterWithServices(cfg, httpapi.Services{
			Documents: documentService,
			Indexing:  indexingService,
			Retrieval: retrievalService,
			RAG:       ragService,
			Traces:    tracingRepository,
		}),
	}

	log.Printf("starting %s server on %s", cfg.App.Name, cfg.Server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
