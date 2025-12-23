package openai

import (
	"context"

	"github.com/yoga/knowledge-base/internal/domain/ai"
	aiservice "github.com/yoga/knowledge-base/internal/service/ai"
)

// Adapter OpenAI客户端适配器，实现AI服务的OpenAIClient接口
type Adapter struct {
	client *Client
}

// NewAdapter 创建适配器
func NewAdapter(client *Client) *Adapter {
	return &Adapter{client: client}
}

// Chat 实现AI服务的OpenAIClient接口
func (a *Adapter) Chat(ctx context.Context, messages []ai.ChatMessage, tools []aiservice.Tool) (*ai.ChatResponse, error) {
	// 转换工具类型
	openAITools := make([]Tool, len(tools))
	for i, tool := range tools {
		openAITools[i] = Tool{
			Type: tool.Type,
			Function: FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}

	return a.client.Chat(ctx, messages, openAITools)
}
