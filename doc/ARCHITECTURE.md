# 思宁瑜伽知识库系统 - 架构与功能文档

## 项目概述

思宁瑜伽知识库系统是一个基于RAG（检索增强生成）技术的智能问答系统，专为瑜伽馆业务场景设计。系统集成了知识库管理、AI智能问答、课程预订等核心功能，支持微信小程序前端，为瑜伽馆提供全方位的数字化服务。

### 核心特性

- **智能知识库管理**：支持文本、图片、视频等多种内容类型的知识库管理
- **RAG增强问答**：基于向量检索的智能问答，结合知识库内容提供准确回答
- **课程预订系统**：完整的课程管理、预订、评价功能
- **MCP工具集成**：通过MCP协议实现AI与业务系统的无缝集成
- **微信小程序前端**：提供友好的移动端用户体验

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      微信小程序前端                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  首页    │  │  AI问答  │  │  课程预订 │  │  知识库  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │ HTTP/HTTPS
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go API服务层 (Gin)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 知识库服务   │  │  AI问答服务   │  │  课程预订服务│      │
│  │ Knowledge   │  │  AI Service  │  │  Booking     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              MCP Server (业务工具集成)                │  │
│  │  - query_schedule  - book_class                       │  │
│  │  - cancel_booking  - query_user_bookings             │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ PostgreSQL   │  │    MinIO      │  │   Qdrant      │  │  DeepSeek API │
│ (元数据/业务)│  │  (文件存储)   │  │  (向量数据库) │  │  (AI模型)     │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
         │              │              │
         │              │              │
         └──────────────┴──────────────┘
                            │
                            ▼
         ┌──────────────────────────────────────┐
         │      Python服务层 (FastAPI)          │
         │  ┌──────────────┐  ┌──────────────┐ │
         │  │ 向量检索服务 │  │ Embedding服务│ │
         │  │ Vector       │  │ Embedding    │ │
         │  │ Service      │  │ Service      │ │
         │  └──────────────┘  └──────────────┘ │
         └──────────────────────────────────────┘
```

### 技术栈

#### 后端服务（Go）
- **框架**：Gin (HTTP框架)
- **数据库ORM**：原生SQL + database/sql
- **日志**：zap
- **追踪**：OpenTelemetry + Jaeger
- **配置管理**：godotenv

#### Python服务
- **框架**：FastAPI
- **向量化模型**：sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
- **向量数据库客户端**：qdrant-client

#### 存储层
- **PostgreSQL 15**：元数据、业务数据存储
- **MinIO**：文件对象存储（兼容S3）
- **Qdrant**：向量数据库，用于语义搜索

#### AI服务
- **DeepSeek API**：大语言模型（兼容OpenAI API格式）
- **本地Embedding模型**：文本向量化

#### 前端
- **微信小程序**：原生小程序开发

#### 基础设施
- **Docker & Docker Compose**：容器化部署
- **Jaeger**：分布式追踪

## 核心功能模块

### 1. 知识库管理模块

#### 功能概述
提供完整的知识库CRUD操作，支持多种内容类型的知识项管理。

#### 核心功能
- **知识库管理**
  - 创建、查询、更新、删除知识库
  - 知识库元数据管理（名称、描述等）
  
- **知识项管理**
  - 文本知识项：直接创建文本内容
  - 文件知识项：支持上传文件（图片、视频等）
  - 知识项查询、删除
  
- **向量化处理**
  - 自动将知识项内容向量化
  - 存储向量到Qdrant向量数据库
  - 支持增量更新和删除

#### 数据模型
```go
KnowledgeBase {
    ID          UUID
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

KnowledgeItem {
    ID              UUID
    KnowledgeBaseID UUID
    Title           string
    Content         string
    ContentType     string  // text, image, video
    VectorID        string  // Qdrant中的向量ID
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### 关键文件
- `internal/service/knowledge/service.go` - 知识库业务逻辑
- `internal/repository/postgres/knowledge.go` - 数据访问层
- `internal/domain/knowledge/entity.go` - 领域模型

### 2. AI问答模块

#### 功能概述
基于RAG（检索增强生成）技术的智能问答系统，结合知识库内容和业务工具，提供准确的回答。

#### 工作流程
```
用户提问
    ↓
问题向量化 (Embedding服务)
    ↓
向量检索 (Qdrant) → 获取Top-K相关文档
    ↓
构建Prompt (知识库内容 + 用户问题)
    ↓
调用DeepSeek API
    ↓
解析工具调用 (MCP工具)
    ↓
执行业务操作 (如预订课程)
    ↓
返回最终回答
```

#### 核心功能
- **语义检索**：基于向量相似度的知识库检索
- **RAG增强**：将检索到的知识库内容作为上下文
- **工具调用**：自动识别用户意图并调用MCP工具
- **多轮对话**：支持对话历史上下文

#### 关键文件
- `internal/service/ai/service.go` - AI问答核心逻辑
- `internal/service/ai/retriever.go` - 知识库检索
- `pkg/openai/client.go` - DeepSeek API客户端

### 3. 课程预订模块

#### 功能概述
完整的课程管理、预订、评价系统。

#### 核心功能
- **课程管理**
  - 课程创建、查询、更新
  - 课程时间、老师、容量管理
  
- **预订管理**
  - 课程预订
  - 预订查询（用户、课程）
  - 预订取消
  
- **评价系统**
  - 课程评价创建
  - 评价查询（课程、用户）

#### 数据模型
```go
Class {
    ID          UUID
    Name        string
    Instructor  string
    StartTime   time.Time
    EndTime     time.Time
    Capacity    int
    BookedCount int
    Description string
}

Booking {
    ID        UUID
    ClassID   UUID
    UserID    string
    UserName  string
    CreatedAt time.Time
}

Review {
    ID        UUID
    ClassID   UUID
    UserID    string
    UserName  string
    Rating    int
    Content   string
    Images    []string
    CreatedAt time.Time
}
```

#### 关键文件
- `internal/service/booking/service.go` - 预订业务逻辑
- `internal/repository/postgres/booking.go` - 数据访问层
- `internal/domain/booking/entity.go` - 领域模型

### 4. MCP Server模块

#### 功能概述
实现MCP（Model Context Protocol）协议，将业务功能封装为工具，供AI服务调用。

#### 可用工具
- **query_schedule**：查询课程表
  - 支持按日期范围查询
  - 返回课程列表及详细信息
  
- **book_class**：预订课程
  - 参数：课程ID、用户ID、用户名
  - 返回：预订结果
  
- **cancel_booking**：取消预订
  - 参数：预订ID
  - 返回：取消结果
  
- **query_user_bookings**：查询用户预订
  - 参数：用户ID
  - 返回：用户的所有预订记录

#### 工作流程
```
AI服务识别用户意图
    ↓
调用MCP工具
    ↓
MCP Server执行业务逻辑
    ↓
返回执行结果
    ↓
AI服务整合结果到回答中
```

#### 关键文件
- `pkg/mcp/server.go` - MCP协议实现
- `internal/mcp/service.go` - MCP服务封装
- `internal/mcp/tools/booking.go` - 预订工具实现

### 5. 向量检索服务（Python）

#### 功能概述
提供向量存储和相似度检索功能。

#### 核心功能
- **向量存储**：将知识库内容的向量存储到Qdrant
- **相似度检索**：根据查询向量检索最相似的知识项
- **集合管理**：管理Qdrant中的向量集合

#### API接口
- `POST /search` - 向量检索
- `POST /store` - 存储向量
- `DELETE /delete/{point_id}` - 删除向量点

#### 关键文件
- `python/vector_service/app.py` - FastAPI应用
- `pkg/vector/client.go` - Go客户端

### 6. Embedding服务（Python）

#### 功能概述
提供文本向量化服务，将文本转换为向量表示。

#### 核心功能
- **文本向量化**：使用本地模型将文本转换为向量
- **批量处理**：支持批量向量化
- **多语言支持**：支持中文等多语言

#### 技术细节
- **模型**：sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
- **向量维度**：384维
- **首次加载**：模型首次使用时自动下载

#### API接口
- `POST /embed` - 文本向量化
- `GET /health` - 健康检查

#### 关键文件
- `python/embedding_service/app.py` - FastAPI应用
- `pkg/embedding/client.go` - Go客户端

### 7. 微信小程序前端

#### 页面结构
- **首页** (`pages/index/`)：课程概览，周一到周日课程列表
- **AI问答** (`pages/chat/`)：智能问答界面，支持多轮对话
- **课程预订** (`pages/booking/`)：课程列表、预订功能
- **课程详情** (`pages/class-detail/`)：课程详细信息、评价
- **知识库** (`pages/knowledge/`)：知识库浏览

#### 核心功能
- **课程展示**：周视图展示，点击日期查看当天课程
- **智能问答**：与AI助手对话，支持课程预订等操作
- **课程预订**：查看课程、预订、取消预订
- **知识库浏览**：浏览知识库内容和知识项

#### 技术特点
- 紫色主题UI设计
- 卡片式布局
- 响应式交互

## 数据流

### 知识库内容上传流程
```
1. 用户上传知识项（文本/文件）
   ↓
2. Go服务接收请求，存储元数据到PostgreSQL
   ↓
3. 文件存储到MinIO（如果是文件类型）
   ↓
4. 调用Embedding服务，将内容向量化
   ↓
5. 调用向量服务，存储向量到Qdrant
   ↓
6. 更新知识项，记录向量ID
```

### AI问答流程
```
1. 用户提问
   ↓
2. Go服务接收问题
   ↓
3. 调用Embedding服务，将问题向量化
   ↓
4. 调用向量服务，在Qdrant中检索相似内容
   ↓
5. 获取Top-K相关文档
   ↓
6. 构建Prompt（知识库内容 + 用户问题）
   ↓
7. 调用DeepSeek API
   ↓
8. 解析响应，检查是否有工具调用
   ↓
9. 如有工具调用，执行MCP工具
   ↓
10. 整合结果，返回最终回答
```

### 课程预订流程
```
1. 用户选择课程并预订
   ↓
2. Go服务接收预订请求
   ↓
3. 检查课程容量
   ↓
4. 创建预订记录（PostgreSQL）
   ↓
5. 更新课程已预订数量
   ↓
6. 返回预订结果
```

## API接口

### 知识库API

#### 知识库管理
- `POST /api/v1/knowledge-bases` - 创建知识库
- `GET /api/v1/knowledge-bases` - 列出知识库
- `GET /api/v1/knowledge-bases/:id` - 获取知识库
- `PUT /api/v1/knowledge-bases/:id` - 更新知识库
- `DELETE /api/v1/knowledge-bases/:id` - 删除知识库

#### 知识项管理
- `POST /api/v1/knowledge-bases/:base_id/items/text` - 创建文本知识项
- `POST /api/v1/knowledge-bases/:base_id/items/file` - 上传文件知识项
- `GET /api/v1/knowledge-bases/:base_id/items` - 列出知识项
- `GET /api/v1/knowledge-bases/:base_id/items/:id` - 获取知识项
- `DELETE /api/v1/knowledge-bases/:base_id/items/:id` - 删除知识项

### AI问答API

- `POST /api/v1/ai/chat` - AI聊天

**请求示例**：
```json
{
  "message": "如何预订明天的瑜伽课？",
  "history": [
    {
      "role": "user",
      "content": "你好"
    },
    {
      "role": "assistant",
      "content": "你好！我是AI助手..."
    }
  ],
  "base_id": "optional-knowledge-base-id"
}
```

**响应示例**：
```json
{
  "message": "您可以通过以下方式预订明天的瑜伽课...",
  "sources": [
    {
      "id": "item-id",
      "title": "课程预订指南",
      "content": "预订课程的方法...",
      "score": 0.95
    }
  ],
  "tool_calls": [
    {
      "name": "book_class",
      "arguments": {
        "class_id": "class-uuid",
        "user_id": "user123",
        "user_name": "张三"
      },
      "result": {
        "booking_id": "booking-uuid",
        "status": "success"
      }
    }
  ]
}
```

### 课程和预订API

- `GET /api/v1/classes` - 列出课程（支持日期过滤）
- `GET /api/v1/classes/:id` - 获取课程详情
- `POST /api/v1/classes/:id/book` - 预订课程
- `DELETE /api/v1/classes/bookings/:id` - 取消预订
- `GET /api/v1/classes/bookings` - 查询用户预订
- `POST /api/v1/classes/:id/reviews` - 创建评价
- `GET /api/v1/classes/:id/reviews` - 列出课程评价

## 部署架构

### 服务部署

#### Docker Compose服务
- **PostgreSQL**：端口 5432
- **MinIO**：端口 9000（API），9001（Console）
- **Qdrant**：端口 6333（API），6334（gRPC）
- **Jaeger**：端口 16686（UI），14268（API）

#### Python服务（独立进程）
- **向量检索服务**：端口 8003
- **Embedding服务**：端口 8002

#### Go服务（独立进程）
- **API服务**：端口 8080

### 部署流程

1. **基础设施启动**：Docker Compose启动数据库、存储等
2. **Python服务启动**：启动向量和Embedding服务
3. **Go服务启动**：启动API服务
4. **健康检查**：验证所有服务正常运行

### 部署脚本

- `deploy.sh`：一键部署脚本，自动停止旧服务并启动新服务
- `stop.sh`：停止所有服务脚本

## 项目结构

```
sining-yoga/
├── cmd/                          # 应用入口
│   └── api-server/              # API服务器主程序
│       └── main.go
├── internal/                     # 内部代码（不对外暴露）
│   ├── api/                     # HTTP处理层
│   │   ├── handler/             # 请求处理器
│   │   │   ├── ai.go           # AI问答处理器
│   │   │   ├── booking.go      # 课程预订处理器
│   │   │   └── knowledge.go    # 知识库处理器
│   │   └── middleware/          # 中间件
│   │       ├── cors.go         # CORS支持
│   │       ├── logging.go      # 日志中间件
│   │       └── tracing.go      # 追踪中间件
│   ├── service/                 # 业务逻辑层
│   │   ├── ai/                 # AI问答服务
│   │   │   ├── service.go     # AI服务主逻辑
│   │   │   └── retriever.go   # 知识库检索
│   │   ├── booking/           # 课程预订服务
│   │   │   └── service.go
│   │   └── knowledge/         # 知识库服务
│   │       └── service.go
│   ├── repository/             # 数据访问层
│   │   └── postgres/          # PostgreSQL实现
│   │       ├── booking.go
│   │       ├── db.go
│   │       └── knowledge.go
│   ├── domain/                 # 领域模型
│   │   ├── ai/                # AI领域模型
│   │   ├── booking/           # 预订领域模型
│   │   └── knowledge/         # 知识库领域模型
│   ├── config/                 # 配置管理
│   │   └── config.go
│   └── mcp/                    # MCP服务
│       ├── service.go
│       └── tools/              # MCP工具
│           └── booking.go
├── pkg/                         # 公共包（可被外部使用）
│   ├── storage/                # 存储抽象
│   │   └── storage.go
│   ├── vector/                 # 向量服务客户端
│   │   └── client.go
│   ├── embedding/              # Embedding服务客户端
│   │   └── client.go
│   ├── openai/                 # AI客户端（DeepSeek）
│   │   ├── client.go
│   │   └── adapter.go
│   ├── mcp/                    # MCP协议
│   │   ├── protocol.go
│   │   └── server.go
│   └── observability/          # 可观测性
│       └── tracer.go
├── python/                     # Python服务
│   ├── vector_service/        # 向量检索服务
│   │   ├── app.py
│   │   ├── requirements.txt
│   │   └── Dockerfile
│   └── embedding_service/     # Embedding服务
│       ├── app.py
│       ├── requirements.txt
│       └── Dockerfile
├── miniprogram/                # 微信小程序
│   ├── app.js                 # 小程序入口
│   ├── app.json               # 小程序配置
│   ├── pages/                 # 页面
│   │   ├── index/             # 首页
│   │   ├── chat/              # AI问答
│   │   ├── booking/           # 课程预订
│   │   ├── class-detail/      # 课程详情
│   │   └── knowledge/         # 知识库
│   └── utils/                 # 工具函数
│       └── api.js             # API调用封装
├── configs/                    # 配置文件
│   └── init.sql               # 数据库初始化脚本
├── doc/                        # 文档
│   ├── ARCHITECTURE.md        # 本文档
│   ├── MINIPROGRAM_SETUP.md   # 小程序配置指南
│   └── ...
├── docker-compose.yml          # Docker编排配置
├── deploy.sh                   # 部署脚本
├── stop.sh                     # 停止脚本
├── Makefile                    # Make命令
├── go.mod                      # Go模块定义
└── README.md                   # 项目说明
```

## 配置说明

### 环境变量

#### 服务器配置
- `SERVER_HOST`：服务器监听地址（默认：0.0.0.0）
- `SERVER_PORT`：服务器端口（默认：8080）

#### 数据库配置
- `DB_HOST`：PostgreSQL主机（默认：localhost）
- `DB_PORT`：PostgreSQL端口（默认：5432）
- `DB_USER`：数据库用户（默认：yoga）
- `DB_PASSWORD`：数据库密码（默认：yoga123）
- `DB_NAME`：数据库名称（默认：yoga_db）

#### MinIO配置
- `MINIO_ENDPOINT`：MinIO端点（默认：localhost:9000）
- `MINIO_ACCESS_KEY_ID`：访问密钥ID
- `MINIO_SECRET_ACCESS_KEY`：访问密钥
- `MINIO_BUCKET_NAME`：存储桶名称（默认：yoga-knowledge）

#### Qdrant配置
- `QDRANT_HOST`：Qdrant主机（默认：localhost）
- `QDRANT_PORT`：Qdrant端口（默认：6333）

#### AI配置
- `OPENAI_API_KEY`：DeepSeek API密钥（必填）
- `OPENAI_BASE_URL`：API基础URL（默认：https://api.deepseek.com/v1）
- `OPENAI_MODEL`：模型名称（默认：deepseek-chat）

#### 服务URL配置
- `VECTOR_SERVICE_URL`：向量服务URL（默认：http://localhost:8003）
- `EMBEDDING_SERVICE_URL`：Embedding服务URL（默认：http://localhost:8002）

#### 追踪配置
- `JAEGER_ENDPOINT`：Jaeger端点（默认：http://localhost:14268/api/traces）

## 扩展开发

### 添加新的MCP工具

1. 在 `internal/mcp/tools/` 中创建工具文件
2. 实现工具函数
3. 使用 `RegisterXXXTools` 函数注册工具
4. 在 `cmd/api-server/main.go` 中调用注册函数

### 扩展知识库类型

1. 在 `internal/domain/knowledge/entity.go` 中添加新的内容类型
2. 在 `internal/service/knowledge/service.go` 中实现处理逻辑
3. 更新向量化服务以支持新类型

### 添加新的API端点

1. 在 `internal/api/handler/` 中添加处理器
2. 在 `cmd/api-server/main.go` 中注册路由
3. 实现相应的业务逻辑

## 性能优化

### 向量检索优化
- 使用Qdrant的索引优化
- 调整检索参数（Top-K数量）
- 缓存常用查询结果

### 数据库优化
- 添加适当的索引
- 使用连接池
- 优化查询语句

### 服务优化
- 使用异步处理（如向量化任务）
- 实现请求限流
- 添加缓存层

## 监控与运维

### 日志
- 使用zap进行结构化日志记录
- 日志级别可配置
- 日志输出到文件和控制台

### 追踪
- 使用OpenTelemetry进行分布式追踪
- Jaeger UI查看追踪信息
- 追踪所有关键操作

### 健康检查
- `/health` 端点提供健康检查
- Docker健康检查配置
- 服务启动时验证依赖服务

## 安全考虑

### API安全
- CORS配置（允许小程序跨域）
- 输入验证
- 错误信息不泄露敏感信息

### 数据安全
- 数据库连接使用SSL（生产环境）
- MinIO访问控制
- API密钥安全存储

## 未来规划

### 功能扩展
- 用户认证和授权
- 多租户支持
- 更丰富的知识库内容类型
- 实时通知功能

### 技术优化
- 引入消息队列（异步任务处理）
- 实现服务发现和负载均衡
- 容器化Python服务
- 性能监控和告警

### 用户体验
- 小程序UI优化
- 更多交互功能
- 离线支持

---

**文档版本**：v1.0  
**最后更新**：2024-12-23  
**维护者**：思宁瑜伽技术团队

