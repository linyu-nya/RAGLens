# RAGLens

[中文文档](README.zh-CN.md)

RAGLens is a local RAG debugging workbench. It provides document upload, chunking, embedding indexing, vector retrieval, RAG asking, and trace inspection.

## Stack

- Backend: Go, pgx, PostgreSQL, pgvector
- Frontend: React, Vite, TanStack Query, TanStack Router, CodeMirror
- Database: PostgreSQL 16 with pgvector

## Requirements

- Go 1.26+
- Node.js 20+
- Docker Engine with Docker Compose
- An OpenAI-compatible model API key

## Quick Start

Create a local environment file:

```bash
cp .env.example .env
```

Edit `.env` and set `RAGLENS_OPENAI_API_KEY`. Then run:

```bash
npm install
npm run dev
```

`npm run dev` will:

- start PostgreSQL with Docker Compose
- wait for the database port
- run database migrations through goose
- start the Go backend on `http://localhost:8080`
- start the Vite frontend on `http://127.0.0.1:5173`

Open the app:

```text
http://127.0.0.1:5173
```

If your Docker Engine is inside WSL, run the commands from WSL in this project directory.

## Model Configuration

OpenAI-compatible default:

```dotenv
RAGLENS_OPENAI_API_KEY=<your-api-key>
```

Example for GLM / BigModel:

```dotenv
RAGLENS_OPENAI_BASE_URL=https://open.bigmodel.cn/api/paas/v4
RAGLENS_OPENAI_API_KEY=<your-api-key>
RAGLENS_CHAT_MODEL=<your-chat-model>
RAGLENS_EMBEDDING_MODEL=embedding-3
RAGLENS_EMBEDDING_DIMENSIONS=2048
```

Other optional backend settings:

```dotenv
RAGLENS_SERVER_ADDR=:8080
RAGLENS_DATABASE_URL=postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable
```

## Manual Commands

Start PostgreSQL:

```bash
docker compose up -d postgres
docker compose ps
```

Run migrations:

```bash
npm run db:migrate
```

Run backend only:

```bash
cd raglens-server
go run ./cmd/raglens-server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

Run frontend only:

```bash
npm run dev --workspace raglens-web
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
npm run build:server
```

Frontend:

```bash
npm run build:web
```

Both:

```bash
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
