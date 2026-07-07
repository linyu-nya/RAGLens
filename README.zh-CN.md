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
- OpenAI-compatible 模型 API Key

## 快速开始

创建本地环境变量文件：

```bash
cp .env.example .env
```

编辑 `.env`，填入 `RAGLENS_OPENAI_API_KEY`。然后运行：

```bash
npm install
npm run dev
```

`npm run dev` 会自动完成：

- 使用 Docker Compose 启动 PostgreSQL
- 等待数据库端口可用
- 通过 goose 执行数据库迁移
- 启动 Go 后端：`http://localhost:8080`
- 启动 Vite 前端：`http://127.0.0.1:5173`

打开：

```text
http://127.0.0.1:5173
```

如果你的 Docker Engine 在 WSL 中，请在 WSL 里的项目目录中运行上述命令。

## 模型配置

OpenAI-compatible 默认配置：

```dotenv
RAGLENS_OPENAI_API_KEY=<your-api-key>
```

使用 GLM / BigModel 的示例：

```dotenv
RAGLENS_OPENAI_BASE_URL=https://open.bigmodel.cn/api/paas/v4
RAGLENS_OPENAI_API_KEY=<your-api-key>
RAGLENS_CHAT_MODEL=<your-chat-model>
RAGLENS_EMBEDDING_MODEL=embedding-3
RAGLENS_EMBEDDING_DIMENSIONS=2048
```

其他可选后端配置：

```dotenv
RAGLENS_SERVER_ADDR=:8080
RAGLENS_DATABASE_URL=postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable
```

## 手动命令

启动 PostgreSQL：

```bash
docker compose up -d postgres
docker compose ps
```

执行数据库迁移：

```bash
npm run db:migrate
```

只启动后端：

```bash
cd raglens-server
go run ./cmd/raglens-server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

只启动前端：

```bash
npm run dev --workspace raglens-web
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
npm run build:server
```

前端：

```bash
npm run build:web
```

全部构建：

```bash
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
