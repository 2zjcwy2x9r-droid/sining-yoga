#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# 创建必要的目录
mkdir -p logs bin

echo -e "${GREEN}========================================"
echo -e "${GREEN}开始部署瑜伽知识库系统"
echo -e "${GREEN}========================================"

# 1. 停止并清理所有相关进程
echo -e "${YELLOW}[1/6] 停止所有相关进程...${NC}"

# 停止Docker容器 - 更彻底的清理
echo "  停止Docker容器..."
# 先停止所有相关容器（包括已停止的）
for container in yoga-postgres yoga-minio yoga-qdrant yoga-jaeger yoga-vector-service yoga-embedding-service; do
    docker stop "$container" 2>/dev/null || true
    docker rm -f "$container" 2>/dev/null || true
done
# 停止所有docker-compose服务并清理
docker-compose down -v --remove-orphans 2>/dev/null || true
# 强制删除所有相关容器（包括已停止的）
docker ps -a --filter "name=yoga-" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
# 等待端口释放
sleep 3

# 杀死Go API服务进程
echo "  停止Go API服务..."
pkill -f "api-server" || true
pkill -f "go run.*main.go" || true
sleep 2

# 杀死Python服务进程
echo "  停止Python服务..."
pkill -f "vector_service" || true
pkill -f "embedding_service" || true
pkill -f "uvicorn.*vector_service" || true
pkill -f "uvicorn.*embedding_service" || true
sleep 2

# 清理端口占用（强制清理）
echo "  清理端口占用..."
# 清理端口占用的进程（强制kill）
for port in 8080 8003 8002 5432 9000 6333 16686 14268; do
    pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "    强制清理端口 $port 的占用进程..."
        echo $pids | xargs kill -9 2>/dev/null || true
    fi
done
# 再次查找并清理所有相关容器（通过名称匹配）
echo "  清理所有相关容器..."
docker ps -a --filter "name=yoga-" --format "{{.ID}}" | while read container_id; do
    if [ -n "$container_id" ]; then
        docker stop "$container_id" 2>/dev/null || true
        docker rm -f "$container_id" 2>/dev/null || true
    fi
done
# 等待端口完全释放
sleep 3

echo -e "${GREEN}✓ 所有进程已停止${NC}"

# 2. 检查环境
echo -e "${YELLOW}[2/6] 检查环境...${NC}"

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker未安装${NC}"
    exit 1
fi

# 检查Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}错误: Docker Compose未安装${NC}"
    exit 1
fi

# 检查Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go未安装${NC}"
    exit 1
fi

# 检查Python
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}错误: Python3未安装${NC}"
    exit 1
fi

# 检查环境变量文件并加载
if [ -f ".env" ]; then
    echo "  加载环境变量文件..."
    set -a
    source .env
    set +a
    echo -e "${GREEN}✓ 环境变量已加载${NC}"
else
    echo -e "${YELLOW}警告: .env文件不存在，将使用默认配置${NC}"
    echo -e "${YELLOW}提示: Embedding服务使用本地模型，无需API Key${NC}"
fi

# 检查关键环境变量（AI服务需要OPENAI_API_KEY，但Embedding服务使用本地模型）
if [ -z "$OPENAI_API_KEY" ]; then
    # 尝试从系统环境变量获取
    if [ -n "${OPENAI_API_KEY:-}" ]; then
        export OPENAI_API_KEY
    else
        echo -e "${YELLOW}警告: OPENAI_API_KEY未设置，AI聊天服务可能无法使用${NC}"
        echo -e "${YELLOW}提示: 请在.env文件中设置OPENAI_API_KEY（DeepSeek API Key），或通过环境变量导出${NC}"
        echo -e "${YELLOW}注意: Embedding服务使用本地模型，不受此影响${NC}"
    fi
fi

echo -e "${GREEN}✓ 环境检查通过${NC}"

# 3. 启动基础设施服务
echo -e "${YELLOW}[3/6] 启动基础设施服务（Docker）...${NC}"

# 再次确认清理（防止残留容器）
echo "  最终确认清理残留容器..."
docker ps -a --filter "name=yoga-" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
sleep 1

# 启动Docker Compose服务（只启动基础设施服务，不启动vector-service和embedding-service）
docker-compose up -d postgres minio qdrant jaeger

# 等待服务就绪
echo "  等待PostgreSQL就绪..."
timeout=60
counter=0
until docker-compose exec -T postgres pg_isready -U yoga &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}错误: PostgreSQL启动超时${NC}"
        exit 1
    fi
done

echo "  等待MinIO就绪..."
counter=0
until curl -s http://localhost:9000/minio/health/live &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}错误: MinIO启动超时${NC}"
        exit 1
    fi
done

echo "  等待Qdrant就绪..."
counter=0
until curl -s http://localhost:6333/health &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}错误: Qdrant启动超时${NC}"
        exit 1
    fi
done

echo -e "${GREEN}✓ 基础设施服务已启动${NC}"

# 4. 安装Python依赖
echo -e "${YELLOW}[4/6] 安装Python依赖...${NC}"

# 检查并创建虚拟环境
if [ ! -d "venv" ]; then
    echo "  创建Python虚拟环境..."
    python3 -m venv venv
fi

# 激活虚拟环境
source venv/bin/activate

# 安装依赖
echo "  安装向量服务依赖..."
cd python/vector_service
pip install -q -r requirements.txt
cd "$PROJECT_ROOT"

echo "  安装Embedding服务依赖..."
cd python/embedding_service
pip install -r requirements.txt
cd "$PROJECT_ROOT"

echo -e "${GREEN}✓ Python依赖已安装${NC}"

# 5. 启动Python服务
echo -e "${YELLOW}[5/6] 启动Python服务...${NC}"

# 启动向量服务
echo "  启动向量检索服务..."
# 检查端口是否可用
if lsof -ti:8003 >/dev/null 2>&1; then
    echo -e "${YELLOW}警告: 端口8003被占用，正在清理...${NC}"
    # 停止可能占用端口的容器
    docker stop $(docker ps -q --filter "publish=8003" 2>/dev/null || true) 2>/dev/null || true
    docker rm $(docker ps -q --filter "publish=8003" 2>/dev/null || true) 2>/dev/null || true
    # 清理端口占用的进程
    lsof -ti:8003 | xargs kill -9 2>/dev/null || true
    sleep 2
fi
cd python/vector_service
nohup python -m uvicorn app:app --host 0.0.0.0 --port 8003 > "$PROJECT_ROOT/logs/vector_service.log" 2>&1 &
VECTOR_PID=$!
cd "$PROJECT_ROOT"

# 等待向量服务就绪
sleep 3
counter=0
until curl -s http://localhost:8003/health &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    # 检查进程是否还在运行
    if ! kill -0 $VECTOR_PID 2>/dev/null; then
        echo -e "${RED}错误: 向量服务进程已退出，启动失败${NC}"
        echo "查看日志: tail -f logs/vector_service.log"
        exit 1
    fi
    if [ $counter -ge 30 ]; then
        echo -e "${RED}错误: 向量服务启动超时${NC}"
        echo "查看日志: tail -f logs/vector_service.log"
        exit 1
    fi
done

# 启动Embedding服务
echo "  启动Embedding服务..."
echo "  注意: Embedding服务使用本地模型，无需API Key"
# 检查端口是否可用
if lsof -ti:8002 >/dev/null 2>&1; then
    echo -e "${YELLOW}警告: 端口8002被占用，正在清理...${NC}"
    # 停止可能占用端口的容器
    docker stop $(docker ps -q --filter "publish=8002" 2>/dev/null || true) 2>/dev/null || true
    docker rm $(docker ps -q --filter "publish=8002" 2>/dev/null || true) 2>/dev/null || true
    # 清理端口占用的进程
    lsof -ti:8002 | xargs kill -9 2>/dev/null || true
    sleep 2
fi
cd python/embedding_service
# 使用本地模型，传递模型名称（如果指定）
nohup env EMBEDDING_MODEL="${EMBEDDING_MODEL:-sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2}" python -m uvicorn app:app --host 0.0.0.0 --port 8002 > "$PROJECT_ROOT/logs/embedding_service.log" 2>&1 &
EMBEDDING_PID=$!
cd "$PROJECT_ROOT"

# 等待Embedding服务就绪
# Embedding服务可能需要加载模型，需要更长的启动时间
echo "  等待Embedding服务就绪（最多等待300秒）..."
sleep 5
counter=0
timeout=300
until curl -s http://localhost:8002/health &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    # 每10秒显示一次进度
    if [ $((counter % 10)) -eq 0 ]; then
        echo "    已等待 ${counter} 秒..."
    fi
    # 检查进程是否还在运行
    if ! kill -0 $EMBEDDING_PID 2>/dev/null; then
        echo -e "${RED}错误: Embedding服务进程已退出，启动失败${NC}"
        echo "查看日志: tail -f logs/embedding_service.log"
        exit 1
    fi
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}错误: Embedding服务启动超时（已等待 ${timeout} 秒）${NC}"
        echo "查看日志: tail -f logs/embedding_service.log"
        echo "检查进程状态: ps aux | grep embedding_service"
        exit 1
    fi
done

echo -e "${GREEN}✓ Python服务已启动 (向量服务PID: $VECTOR_PID, Embedding服务PID: $EMBEDDING_PID)${NC}"

# 6. 启动Go API服务
echo -e "${YELLOW}[6/6] 启动Go API服务...${NC}"

# 创建必要的目录
mkdir -p logs bin

# 编译Go程序（可选，也可以直接运行）
echo "  编译Go程序..."
go build -o bin/api-server ./cmd/api-server

# 启动API服务
echo "  启动API服务..."
nohup ./bin/api-server > logs/api_server.log 2>&1 &
API_PID=$!

# 等待API服务就绪
sleep 3
counter=0
until curl -s http://localhost:8080/health &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge 30 ]; then
        echo -e "${RED}错误: API服务启动超时${NC}"
        echo "查看日志: tail -f logs/api_server.log"
        exit 1
    fi
done

echo -e "${GREEN}✓ Go API服务已启动 (PID: $API_PID)${NC}"

# 保存PID到文件
echo "$VECTOR_PID" > logs/vector_service.pid
echo "$EMBEDDING_PID" > logs/embedding_service.pid
echo "$API_PID" > logs/api_server.pid

# 部署完成
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务状态："
echo "  - PostgreSQL:    http://localhost:5432"
echo "  - MinIO:         http://localhost:9000 (Console: http://localhost:9001)"
echo "  - Qdrant:        http://localhost:6333"
echo "  - Jaeger:        http://localhost:16686"
echo "  - 向量服务:      http://localhost:8003"
echo "  - Embedding服务: http://localhost:8002"
echo "  - API服务:       http://localhost:8080"
echo ""
echo "查看日志："
echo "  - API服务:       tail -f logs/api_server.log"
echo "  - 向量服务:      tail -f logs/vector_service.log"
echo "  - Embedding服务: tail -f logs/embedding_service.log"
echo ""
echo "停止服务："
echo "  ./stop.sh"
echo ""

