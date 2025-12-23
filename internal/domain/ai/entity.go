package ai

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message    string   `json:"message"`
	History    []ChatMessage `json:"history,omitempty"`
	BaseID     string   `json:"base_id,omitempty"` // 指定知识库ID
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Message     string   `json:"message"`
	Sources     []Source `json:"sources,omitempty"` // 引用的知识库内容
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"` // MCP工具调用
}

// Source 知识库来源
type Source struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Score   float64 `json:"score"`
}

// ToolCall MCP工具调用
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Result    interface{}            `json:"result,omitempty"`
}

