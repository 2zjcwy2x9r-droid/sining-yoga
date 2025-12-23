.PHONY: deploy stop clean build test help

# 默认目标
.DEFAULT_GOAL := help

# 部署
deploy:
	@./deploy.sh

# 停止服务
stop:
	@./stop.sh

# 清理
clean:
	@echo "清理构建产物和日志..."
	@rm -rf bin/ logs/*.log logs/*.pid
	@docker-compose down -v
	@echo "清理完成"

# 构建
build:
	@echo "构建Go程序..."
	@mkdir -p bin
	@go build -o bin/api-server ./cmd/api-server
	@echo "构建完成: bin/api-server"

# 测试
test:
	@echo "运行测试..."
	@go test ./... -v

# 格式化代码
fmt:
	@echo "格式化Go代码..."
	@go fmt ./...
	@echo "格式化完成"

# 检查代码
lint:
	@echo "检查代码..."
	@golangci-lint run ./... || echo "请安装golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# 安装依赖
deps:
	@echo "安装Go依赖..."
	@go mod download
	@go mod tidy
	@echo "Go依赖安装完成"
	@echo "安装Python依赖..."
	@if [ ! -d "venv" ]; then python3 -m venv venv; fi
	@source venv/bin/activate && \
		pip install -r python/vector_service/requirements.txt && \
		pip install -r python/embedding_service/requirements.txt
	@echo "Python依赖安装完成"

# 查看日志
logs:
	@tail -f logs/api_server.log

# 查看所有日志
logs-all:
	@tail -f logs/*.log

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make deploy      - 部署所有服务"
	@echo "  make stop        - 停止所有服务"
	@echo "  make clean       - 清理构建产物和日志"
	@echo "  make build       - 构建Go程序"
	@echo "  make test        - 运行测试"
	@echo "  make fmt         - 格式化代码"
	@echo "  make lint        - 检查代码"
	@echo "  make deps        - 安装所有依赖"
	@echo "  make logs        - 查看API服务日志"
	@echo "  make logs-all    - 查看所有日志"
	@echo "  make help        - 显示此帮助信息"

