package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yoga/knowledge-base/internal/domain/ai"
)

// Client OpenAI客户端
type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewClient 创建OpenAI客户端
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Tool MCP工具定义（与AI服务Tool兼容）
type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 函数定义
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToAIServiceTool 转换为AI服务的Tool类型
func (t Tool) ToAIServiceTool() interface{} {
	// 返回一个可以转换为AI服务Tool的结构
	return struct {
		Type     string             `json:"type"`
		Function FunctionDefinition `json:"function"`
	}{
		Type:     t.Type,
		Function: t.Function,
	}
}

// ChatRequest OpenAI聊天请求
type ChatRequest struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Tools      []Tool    `json:"tools,omitempty"`
	ToolChoice string    `json:"tool_choice,omitempty"` // "auto", "none", or specific tool
}

// Message 消息
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ChatResponse OpenAI聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

// Choice 选择
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, messages []ai.ChatMessage, tools []Tool) (*ai.ChatResponse, error) {
	// 转换消息格式
	openAIMessages := make([]Message, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 工具已经是正确的格式，直接使用
	openAITools := tools

	reqBody := ChatRequest{
		Model:      c.model,
		Messages:   openAIMessages,
		Tools:      openAITools,
		ToolChoice: "auto",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		
		// 处理常见的错误状态码，提供更友好的错误信息
		var errorMsg string
		switch resp.StatusCode {
		case 401:
			errorMsg = "API密钥无效，请检查OPENAI_API_KEY配置"
		case 402:
			errorMsg = "账户余额不足，请前往DeepSeek平台充值"
		case 429:
			errorMsg = "请求频率过高，请稍后重试"
		case 500, 502, 503:
			errorMsg = "AI服务暂时不可用，请稍后重试"
		default:
			errorMsg = fmt.Sprintf("请求失败，状态码: %d", resp.StatusCode)
		}
		
		return nil, fmt.Errorf("%s (响应: %s)", errorMsg, string(bodyBytes))
	}

	var openAIResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("响应中没有选择")
	}

	choice := openAIResp.Choices[0]
	result := &ai.ChatResponse{
		Message: choice.Message.Content,
	}

	// 处理工具调用
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ai.ToolCall, len(choice.Message.ToolCalls))
		for i, toolCall := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				args = make(map[string]interface{})
			}

			result.ToolCalls[i] = ai.ToolCall{
				Name:      toolCall.Function.Name,
				Arguments: args,
			}
		}
	}

	return result, nil
}
