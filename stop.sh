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

echo -e "${YELLOW}停止所有服务...${NC}"

# 从PID文件停止服务
if [ -f "logs/api_server.pid" ]; then
    PID=$(cat logs/api_server.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo "  停止API服务 (PID: $PID)..."
        kill $PID 2>/dev/null || true
        rm -f logs/api_server.pid
    fi
fi

if [ -f "logs/vector_service.pid" ]; then
    PID=$(cat logs/vector_service.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo "  停止向量服务 (PID: $PID)..."
        kill $PID 2>/dev/null || true
        rm -f logs/vector_service.pid
    fi
fi

if [ -f "logs/embedding_service.pid" ]; then
    PID=$(cat logs/embedding_service.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo "  停止Embedding服务 (PID: $PID)..."
        kill $PID 2>/dev/null || true
        rm -f logs/embedding_service.pid
    fi
fi

# 强制杀死所有相关进程
echo "  清理残留进程..."
pkill -f "api-server" || true
pkill -f "vector_service" || true
pkill -f "embedding_service" || true
pkill -f "uvicorn.*vector_service" || true
pkill -f "uvicorn.*embedding_service" || true

# 停止Docker容器
echo "  停止Docker容器..."
docker-compose down 2>/dev/null || true

echo -e "${GREEN}✓ 所有服务已停止${NC}"

