# RAGLens

RAGLens is a local RAG debugging workbench. It provides document upload, chunking, embedding indexing, vector retrieval, RAG asking, and trace inspection.

## Stack

- Backend: Go, pgx, PostgreSQL, pgvector
- Frontend: React, Vite, TanStack Query, TanStack Router, CodeMirror
- Database: PostgreSQL 16 with pgvector

## Requirements

- Go 1.26+
- Node.js 20+
- Docker Engine with Docker Compose
- goose for database migrations
- An OpenAI-compatible model API key

## Start PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

The default database URL is:

```text
postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable
```

## Run Migrations

```bash
cd raglens-server
goose -dir migrations postgres "postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable" up
```

## Configure Models

OpenAI-compatible defaults:

```powershell
$env:RAGLENS_OPENAI_API_KEY='<your-api-key>'
```

Example for GLM / BigModel:

```powershell
$env:RAGLENS_OPENAI_BASE_URL='https://open.bigmodel.cn/api/paas/v4'
$env:RAGLENS_OPENAI_API_KEY='<your-api-key>'
$env:RAGLENS_CHAT_MODEL='<your-chat-model>'
$env:RAGLENS_EMBEDDING_MODEL='embedding-3'
$env:RAGLENS_EMBEDDING_DIMENSIONS='2048'
```

Other optional backend settings:

```powershell
$env:RAGLENS_SERVER_ADDR=':8080'
$env:RAGLENS_DATABASE_URL='postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable'
```

## Run Backend

```bash
cd raglens-server
go run ./cmd/raglens-server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

## Run Frontend

```bash
cd raglens-web
npm ci
npm run dev
```

Open:

```text
http://127.0.0.1:5173
```

The Vite dev server proxies `/api` and `/healthz` to `http://localhost:8080`.

## Frontend Routes

- `/` overview
- `/documents` upload documents, inspect chunks, and index embeddings
- `/ask` run retrieval search and RAG ask
- `/traces` inspect RAG traces, prompts, steps, and retrieved chunks

## Build

Backend:

```bash
cd raglens-server
go build ./cmd/raglens-server
```

Frontend:

```bash
cd raglens-web
npm ci
npm run build
```

The frontend build output is written to `raglens-web/dist`.

## API Summary

- `GET /healthz`
- `POST /api/documents/upload`
- `GET /api/documents`
- `GET /api/documents/{id}/chunks`
- `POST /api/documents/{id}/index`
- `POST /api/retrieval/search`
- `POST /api/rag/ask`
- `GET /api/traces`
- `GET /api/traces/{traceId}`
