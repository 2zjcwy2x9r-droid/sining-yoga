package mcp

import (
	"context"
	"fmt"

	aiservice "github.com/yoga/knowledge-base/internal/service/ai"
	"github.com/yoga/knowledge-base/pkg/mcp"
	"go.uber.org/zap"
)

// Service MCP服务实现，用于AI服务集成
type Service struct {
	server mcp.Server
	logger *zap.Logger
}

// NewService 创建MCP服务
func NewService(server mcp.Server, logger *zap.Logger) *Service {
	return &Service{
		server: server,
		logger: logger,
	}
}

// CallTool 调用工具
func (s *Service) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	result, err := s.server.CallTool(name, args)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Data, nil
}

// ListTools 列出所有工具
func (s *Service) ListTools(ctx context.Context) ([]aiservice.Tool, error) {
	mcpTools := s.server.ListTools()

	// 转换为AI服务的Tool格式
	tools := make([]aiservice.Tool, len(mcpTools))
	for i, mcpTool := range mcpTools {
		tools[i] = aiservice.Tool{
			Type: "function",
			Function: aiservice.FunctionDefinition{
				Name:        mcpTool.Name,
				Description: mcpTool.Description,
				Parameters:  mcpTool.Parameters,
			},
		}
	}

	return tools, nil
}

