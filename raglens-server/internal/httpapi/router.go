package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/linyu-nya/RAGLens/raglens-server/internal/config"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/document"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/indexing"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/rag"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/retrieval"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/tracing"
)

const maxMultipartMemory = 32 << 20

type DocumentUploader interface {
	Upload(ctx context.Context, input document.UploadInput) (document.Record, error)
	List(ctx context.Context, input document.ListInput) ([]document.Record, error)
	ListChunks(ctx context.Context, documentID int64) ([]document.ChunkRecord, error)
}

type DocumentIndexer interface {
	IndexDocument(ctx context.Context, input indexing.IndexInput) (indexing.IndexResult, error)
}

type Retriever interface {
	Search(ctx context.Context, input retrieval.SearchInput) (retrieval.SearchResult, error)
}

type RAGAsker interface {
	Ask(ctx context.Context, input rag.AskInput) (rag.AskResult, error)
}

type TraceReader interface {
	List(ctx context.Context, input tracing.ListInput) ([]tracing.TraceSummary, error)
	Get(ctx context.Context, traceID string) (tracing.TraceDetail, error)
}

type Services struct {
	Documents DocumentUploader
	Indexing  DocumentIndexer
	Retrieval Retriever
	RAG       RAGAsker
	Traces    TraceReader
}

func NewRouter(cfg config.Config) http.Handler {
	return NewRouterWithServices(cfg, Services{})
}

func NewRouterWithServices(cfg config.Config, services Services) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"app":    cfg.App.Name,
		})
	})
	mux.HandleFunc("POST /api/documents/upload", uploadDocumentHandler(services.Documents))
	mux.HandleFunc("GET /api/documents", listDocumentsHandler(services.Documents))
	mux.HandleFunc("GET /api/documents/{id}/chunks", listDocumentChunksHandler(services.Documents))
	mux.HandleFunc("POST /api/documents/{id}/index", indexDocumentHandler(services.Indexing))
	mux.HandleFunc("POST /api/retrieval/search", searchRetrievalHandler(services.Retrieval))
	mux.HandleFunc("POST /api/rag/ask", askRAGHandler(services.RAG))
	mux.HandleFunc("GET /api/traces", listTracesHandler(services.Traces))
	mux.HandleFunc("GET /api/traces/{traceId}", getTraceHandler(services.Traces))
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func uploadDocumentHandler(uploader DocumentUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if uploader == nil {
			writeError(w, http.StatusServiceUnavailable, "document service is not configured")
			return
		}
		if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		record, err := uploader.Upload(r.Context(), document.UploadInput{
			Filename: header.Filename,
			Size:     header.Size,
			Content:  file,
		})
		if err != nil {
			writeDocumentError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"id":     record.ID,
			"name":   record.Name,
			"type":   record.Type,
			"size":   record.Size,
			"status": record.Status,
		})
	}
}

func writeDocumentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, document.ErrUnsupportedType):
		writeError(w, http.StatusUnsupportedMediaType, "unsupported document type")
	case errors.Is(err, document.ErrFileTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "file too large")
	default:
		writeError(w, http.StatusInternalServerError, "document upload failed")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func listDocumentsHandler(documents DocumentUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if documents == nil {
			writeError(w, http.StatusServiceUnavailable, "document service is not configured")
			return
		}

		input := document.ListInput{
			Page:     intQuery(r, "page"),
			PageSize: intQuery(r, "pageSize"),
		}
		records, err := documents.List(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list documents failed")
			return
		}

		items := make([]map[string]any, 0, len(records))
		for _, record := range records {
			items = append(items, map[string]any{
				"id":     record.ID,
				"name":   record.Name,
				"type":   record.Type,
				"size":   record.Size,
				"status": record.Status,
			})
		}

		page := input.Page
		if page < 1 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize < 1 {
			pageSize = document.DefaultPageSize
		}
		if pageSize > document.MaxPageSize {
			pageSize = document.MaxPageSize
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items":    items,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func intQuery(r *http.Request, key string) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func listDocumentChunksHandler(documents DocumentUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if documents == nil {
			writeError(w, http.StatusServiceUnavailable, "document service is not configured")
			return
		}

		documentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || documentID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid document id")
			return
		}

		chunks, err := documents.ListChunks(r.Context(), documentID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list document chunks failed")
			return
		}

		items := make([]map[string]any, 0, len(chunks))
		for _, chunk := range chunks {
			items = append(items, map[string]any{
				"id":            chunk.ID,
				"documentId":    chunk.DocumentID,
				"chunkIndex":    chunk.ChunkIndex,
				"content":       chunk.Content,
				"contentLength": chunk.ContentLength,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func indexDocumentHandler(indexer DocumentIndexer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if indexer == nil {
			writeError(w, http.StatusServiceUnavailable, "indexing service is not configured")
			return
		}

		documentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || documentID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid document id")
			return
		}

		result, err := indexer.IndexDocument(r.Context(), indexing.IndexInput{DocumentID: documentID})
		if err != nil {
			if errors.Is(err, indexing.ErrInvalidDocumentID) {
				writeError(w, http.StatusBadRequest, "invalid document id")
				return
			}
			writeError(w, http.StatusInternalServerError, "index document failed")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"documentId": result.DocumentID,
			"chunkCount": result.ChunkCount,
		})
	}
}

func searchRetrievalHandler(retriever Retriever) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if retriever == nil {
			writeError(w, http.StatusServiceUnavailable, "retrieval service is not configured")
			return
		}

		var request struct {
			Query string `json:"query"`
			TopK  int    `json:"topK"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		result, err := retriever.Search(r.Context(), retrieval.SearchInput{
			Query: request.Query,
			TopK:  request.TopK,
		})
		if err != nil {
			if errors.Is(err, retrieval.ErrEmptyQuery) {
				writeError(w, http.StatusBadRequest, "query is required")
				return
			}
			writeError(w, http.StatusInternalServerError, "retrieval search failed")
			return
		}

		items := make([]map[string]any, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, map[string]any{
				"chunkId":       item.ChunkID,
				"documentId":    item.DocumentID,
				"documentName":  item.DocumentName,
				"chunkIndex":    item.ChunkIndex,
				"content":       item.Content,
				"contentLength": item.ContentLength,
				"score":         item.Score,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"query": result.Query,
			"items": items,
		})
	}
}

func askRAGHandler(asker RAGAsker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if asker == nil {
			writeError(w, http.StatusServiceUnavailable, "rag service is not configured")
			return
		}

		var request struct {
			Question string `json:"question"`
			TopK     int    `json:"topK"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		result, err := asker.Ask(r.Context(), rag.AskInput{
			Question: request.Question,
			TopK:     request.TopK,
		})
		if err != nil {
			if errors.Is(err, rag.ErrEmptyQuestion) {
				writeError(w, http.StatusBadRequest, "question is required")
				return
			}
			writeError(w, http.StatusInternalServerError, "rag ask failed")
			return
		}

		chunks := make([]map[string]any, 0, len(result.Chunks))
		for _, chunk := range result.Chunks {
			chunks = append(chunks, map[string]any{
				"chunkId":       chunk.ChunkID,
				"documentId":    chunk.DocumentID,
				"documentName":  chunk.DocumentName,
				"chunkIndex":    chunk.ChunkIndex,
				"content":       chunk.Content,
				"contentLength": chunk.ContentLength,
				"score":         chunk.Score,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"traceId":  result.TraceID,
			"question": result.Question,
			"answer":   result.Answer,
			"model":    result.Model,
			"chunks":   chunks,
		})
	}
}

func listTracesHandler(reader TraceReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if reader == nil {
			writeError(w, http.StatusServiceUnavailable, "trace service is not configured")
			return
		}

		items, err := reader.List(r.Context(), tracing.ListInput{
			Page:     intQuery(r, "page"),
			PageSize: intQuery(r, "pageSize"),
			Status:   r.URL.Query().Get("status"),
			Keyword:  r.URL.Query().Get("keyword"),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list traces failed")
			return
		}

		responseItems := make([]map[string]any, 0, len(items))
		for _, item := range items {
			responseItems = append(responseItems, map[string]any{
				"traceId":   item.TraceID,
				"question":  item.Question,
				"answer":    item.Answer,
				"modelName": item.ModelName,
				"status":    item.Status,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": responseItems})
	}
}

func getTraceHandler(reader TraceReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if reader == nil {
			writeError(w, http.StatusServiceUnavailable, "trace service is not configured")
			return
		}

		traceID := r.PathValue("traceId")
		if traceID == "" {
			writeError(w, http.StatusBadRequest, "trace id is required")
			return
		}

		detail, err := reader.Get(r.Context(), traceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "get trace failed")
			return
		}

		steps := make([]map[string]any, 0, len(detail.Steps))
		for _, step := range detail.Steps {
			steps = append(steps, map[string]any{
				"stepName":   step.StepName,
				"stepOrder":  step.StepOrder,
				"inputData":  json.RawMessage(step.InputData),
				"outputData": json.RawMessage(step.OutputData),
				"status":     step.Status,
			})
		}

		chunks := make([]map[string]any, 0, len(detail.RetrievedChunks))
		for _, chunk := range detail.RetrievedChunks {
			chunks = append(chunks, map[string]any{
				"chunkId":         chunk.ChunkID,
				"rank":            chunk.Rank,
				"score":           chunk.Score,
				"contentSnapshot": chunk.ContentSnapshot,
				"documentName":    chunk.DocumentName,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"traceId":         detail.TraceID,
			"question":        detail.Question,
			"finalPrompt":     detail.FinalPrompt,
			"answer":          detail.Answer,
			"modelName":       detail.ModelName,
			"status":          detail.Status,
			"steps":           steps,
			"retrievedChunks": chunks,
		})
	}
}
