# RAGLens

[English README](README.md)

RAGLens 是一个本地 RAG 调试工作台，用于完成文档上传、文本切分、Embedding 索引、向量检索、RAG 问答和 Trace 查看。

## 技术栈

- 后端：Go、pgx、PostgreSQL、pgvector
- 前端：React、Vite、TanStack Query、TanStack Router、CodeMirror
- 数据库：PostgreSQL 16 + pgvector

## 环境要求

- Go 1.26+
- Node.js 20+
- Docker Engine 和 Docker Compose
- goose，用于执行数据库迁移
- OpenAI-compatible 模型 API Key

## 启动 PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

默认数据库连接：

```text
postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable
```

## 执行数据库迁移

```bash
cd raglens-server
goose -dir migrations postgres "postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable" up
```

## 配置模型

OpenAI-compatible 默认配置只需要设置 API Key：

```powershell
$env:RAGLENS_OPENAI_API_KEY='<your-api-key>'
```

使用 GLM / BigModel 的示例：

```powershell
$env:RAGLENS_OPENAI_BASE_URL='https://open.bigmodel.cn/api/paas/v4'
$env:RAGLENS_OPENAI_API_KEY='<your-api-key>'
$env:RAGLENS_CHAT_MODEL='<your-chat-model>'
$env:RAGLENS_EMBEDDING_MODEL='embedding-3'
$env:RAGLENS_EMBEDDING_DIMENSIONS='2048'
```

其他可选后端配置：

```powershell
$env:RAGLENS_SERVER_ADDR=':8080'
$env:RAGLENS_DATABASE_URL='postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable'
```

## 启动后端

```bash
cd raglens-server
go run ./cmd/raglens-server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

## 启动前端

```bash
cd raglens-web
npm ci
npm run dev
```

浏览器打开：

```text
http://127.0.0.1:5173
```

Vite 开发服务器会把 `/api` 和 `/healthz` 代理到 `http://localhost:8080`。

## 前端页面

- `/`：总览
- `/documents`：上传文档、查看 chunks、写入 embedding
- `/ask`：检索调试和 RAG Ask
- `/traces`：查看 RAG trace、prompt、steps 和 retrieved chunks

## 构建

后端：

```bash
cd raglens-server
go build ./cmd/raglens-server
```

前端：

```bash
cd raglens-web
npm ci
npm run build
```

前端构建产物会输出到 `raglens-web/dist`。

## API 摘要

- `GET /healthz`
- `POST /api/documents/upload`
- `GET /api/documents`
- `GET /api/documents/{id}/chunks`
- `POST /api/documents/{id}/index`
- `POST /api/retrieval/search`
- `POST /api/rag/ask`
- `GET /api/traces`
- `GET /api/traces/{traceId}`
