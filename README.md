# 瑜伽知识库系统

一个基于知识库的AI问答系统，支持微信小程序，集成MCP Server用于业务功能（如定课）。

## 系统架构

- **后端服务（Go）**：知识库管理、AI问答、MCP Server
- **Python服务**：向量检索、文本向量化（使用本地模型）
- **存储**：PostgreSQL（元数据）、MinIO（文件）、Qdrant（向量）
- **前端**：微信小程序

## 快速开始

### 1. 环境要求

- Go 1.21+
- Python 3.11+
- Docker & Docker Compose

### 2. 配置环境变量

复制示例配置文件并修改：

```bash
cp .env.example .env
```

编辑 `.env` 文件，至少需要设置 `OPENAI_API_KEY`（DeepSeek API Key）：

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=yoga
DB_PASSWORD=yoga123
DB_NAME=yoga_db

# MinIO配置
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin123
MINIO_BUCKET_NAME=yoga-knowledge

# Qdrant配置
QDRANT_HOST=localhost
QDRANT_PORT=6333

# AI配置（使用DeepSeek）
OPENAI_API_KEY=your-deepseek-api-key-here
OPENAI_BASE_URL=https://api.deepseek.com/v1
OPENAI_MODEL=deepseek-chat

# Embedding配置（使用本地模型，无需API Key）
EMBEDDING_MODEL=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2

# 服务URL
VECTOR_SERVICE_URL=http://localhost:8003
EMBEDDING_SERVICE_URL=http://localhost:8002

# Jaeger配置
JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

### 3. 启动服务

#### 方式一：使用部署脚本（推荐）

```bash
# 一键部署所有服务（会自动停止旧进程并重新部署）
./deploy.sh

# 停止所有服务
./stop.sh
```

#### 方式二：使用Make命令

```bash
# 部署所有服务
make deploy

# 停止所有服务
make stop

# 查看帮助
make help
```

#### 方式三：手动启动

```bash
# 启动基础设施（数据库、MinIO、Qdrant等）
docker-compose up -d

# 等待服务就绪后，启动Python服务
cd python/vector_service && python -m uvicorn app:app --host 0.0.0.0 --port 8003 &
cd python/embedding_service && python -m uvicorn app:app --host 0.0.0.0 --port 8002 &

# 启动Go API服务
cd cmd/api-server && go run main.go
```

### 4. 微信小程序配置

详细配置说明请参考 [MINIPROGRAM_SETUP.md](./MINIPROGRAM_SETUP.md)

**快速开始**：

1. **获取本机IP地址**：
   ```bash
   # Linux/Mac
   hostname -I | awk '{print $1}'
   # 或
   ip route get 1 | awk '{print $7;exit}'
   
   # Windows
   ipconfig
   # 查找 IPv4 地址（通常是 192.168.x.x）
   ```

2. **修改小程序配置**：
   编辑 `miniprogram/app.js`，修改 `apiBaseUrl`：
   ```javascript
   globalData: {
     userInfo: null,
     // 将 localhost 替换为你的本机IP地址
     apiBaseUrl: 'http://192.168.x.x:8080/api/v1'  // 替换为实际IP
   }
   ```

3. **配置微信开发者工具**：
   - 打开微信开发者工具
   - 点击右上角"详情" -> "本地设置"
   - ✅ 勾选"不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书"

4. **测试连接**：
   - 在小程序中打开"AI问答"页面
   - 发送测试消息，检查是否能正常连接

**注意**：
- 小程序不能使用 `localhost` 或 `127.0.0.1`，必须使用本机IP地址
- 确保手机和电脑在同一WiFi网络
- 生产环境需要配置HTTPS域名（详见 MINIPROGRAM_SETUP.md）

## API文档

### 知识库API

- `POST /api/v1/knowledge-bases` - 创建知识库
- `GET /api/v1/knowledge-bases` - 列出知识库
- `GET /api/v1/knowledge-bases/:id` - 获取知识库
- `PUT /api/v1/knowledge-bases/:id` - 更新知识库
- `DELETE /api/v1/knowledge-bases/:id` - 删除知识库

### 知识项API

- `POST /api/v1/knowledge-bases/:base_id/items/text` - 创建文本知识项
- `POST /api/v1/knowledge-bases/:base_id/items/file` - 上传文件知识项
- `GET /api/v1/knowledge-bases/:base_id/items` - 列出知识项
- `GET /api/v1/knowledge-bases/:base_id/items/:id` - 获取知识项
- `DELETE /api/v1/knowledge-bases/:base_id/items/:id` - 删除知识项

### AI问答API

- `POST /api/v1/ai/chat` - AI聊天

请求体：
```json
{
  "message": "如何预订明天的瑜伽课？",
  "history": [],
  "base_id": "optional-knowledge-base-id"
}
```

响应：
```json
{
  "message": "AI回复内容",
  "sources": [
    {
      "id": "source-id",
      "title": "来源标题",
      "content": "来源内容",
      "score": 0.95
    }
  ],
  "tool_calls": [
    {
      "name": "book_class",
      "arguments": {...},
      "result": {...}
    }
  ]
}
```

## MCP工具

系统集成了以下MCP工具：

- `query_schedule` - 查询课程表
- `book_class` - 预订课程
- `cancel_booking` - 取消预订
- `query_user_bookings` - 查询用户预订

AI问答会自动识别用户意图并调用相应的工具。

## 项目结构

```
yoga/
├── cmd/                    # 应用入口
│   └── api-server/        # API服务器
├── internal/              # 内部代码
│   ├── api/               # HTTP处理
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问
│   ├── domain/            # 领域模型
│   ├── config/            # 配置
│   └── mcp/               # MCP服务
├── pkg/                    # 公共包
│   ├── storage/           # 存储抽象
│   ├── vector/            # 向量服务客户端
│   ├── embedding/         # Embedding服务客户端
│   ├── openai/            # AI客户端（兼容OpenAI API格式，支持DeepSeek等）
│   ├── mcp/               # MCP协议
│   └── observability/     # 可观测性
├── python/                 # Python服务
│   ├── vector_service/    # 向量检索服务
│   └── embedding_service/ # Embedding服务
├── miniprogram/            # 微信小程序
├── configs/                # 配置文件
└── docker-compose.yml      # Docker编排
```

## 开发指南

### 添加新的MCP工具

1. 在 `internal/mcp/tools/` 中创建工具文件
2. 使用 `RegisterXXXTools` 函数注册工具
3. 在 `cmd/api-server/main.go` 中调用注册函数

### 扩展知识库类型

1. 在 `internal/domain/knowledge/entity.go` 中添加新的内容类型
2. 在 `internal/service/knowledge/service.go` 中实现处理逻辑
3. 更新向量化服务以支持新类型

## 测试

```bash
# 运行Go测试
go test ./...

# 运行Python测试
cd python && pytest
```

## 部署

使用Docker Compose进行部署：

```bash
docker-compose up -d
```

## 许可证

MIT

