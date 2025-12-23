package mcp

import (
	"fmt"
	"sync"
)

// DefaultServer 默认MCP服务器实现
type DefaultServer struct {
	tools   map[string]Tool
	handlers map[string]ToolHandler
	mu      sync.RWMutex
}

// NewServer 创建MCP服务器
func NewServer() *DefaultServer {
	return &DefaultServer{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterTool 注册工具
func (s *DefaultServer) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

// ListTools 列出所有工具
func (s *DefaultServer) ListTools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return tools
}

// CallTool 调用工具
func (s *DefaultServer) CallTool(name string, args map[string]interface{}) (*ToolResult, error) {
	s.mu.RLock()
	handler, exists := s.handlers[name]
	s.mu.RUnlock()

	if !exists {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("工具 %s 不存在", name),
		}, fmt.Errorf("工具 %s 不存在", name)
	}

	result, err := handler(args)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &ToolResult{
		Success: true,
		Data:    result,
	}, nil
}

