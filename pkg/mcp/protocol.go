package mcp

// Tool MCP工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall MCP工具调用
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Server MCP服务器接口
type Server interface {
	RegisterTool(tool Tool, handler ToolHandler)
	ListTools() []Tool
	CallTool(name string, args map[string]interface{}) (*ToolResult, error)
}

// ToolHandler 工具处理函数
type ToolHandler func(args map[string]interface{}) (interface{}, error)

