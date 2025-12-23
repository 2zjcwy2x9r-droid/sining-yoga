package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/yoga/knowledge-base/internal/domain/ai"
	"github.com/yoga/knowledge-base/internal/repository/postgres"
	"github.com/yoga/knowledge-base/pkg/observability"
	"go.uber.org/zap"
)

// OpenAI 客户端接口
type OpenAIClient interface {
	Chat(ctx context.Context, messages []ai.ChatMessage, tools []Tool) (*ai.ChatResponse, error)
}

// Retriever 检索器接口（定义在retriever.go中）
type Retriever interface {
	Search(ctx context.Context, query string, limit int, knowledgeBaseID string) ([]ai.Source, error)
}

// MCPService MCP服务接口
type MCPService interface {
	CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
	ListTools(ctx context.Context) ([]Tool, error)
}

// Tool MCP工具定义（与OpenAI Tool兼容）
type Tool struct {
	Type        string                 `json:"type"`
	Function    FunctionDefinition     `json:"function"`
}

// FunctionDefinition 函数定义
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Service AI问答服务
type Service struct {
	openAIClient OpenAIClient
	retriever    Retriever
	mcpSvc       MCPService
	kbRepo       *postgres.KnowledgeRepository
	logger       *zap.Logger
}

// NewService 创建AI问答服务
func NewService(openAIClient OpenAIClient, retriever Retriever, mcpSvc MCPService, kbRepo *postgres.KnowledgeRepository, logger *zap.Logger) *Service {
	return &Service{
		openAIClient: openAIClient,
		retriever:    retriever,
		mcpSvc:       mcpSvc,
		kbRepo:       kbRepo,
		logger:       logger,
	}
}

// Chat 处理聊天请求
func (s *Service) Chat(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	ctx, span := observability.StartSpan(ctx, "ai-service", "Chat")
	defer span.End()

	// 1. 从知识库检索相关内容
	var sources []ai.Source
	if req.Message != "" {
		var err error
		sources, err = s.retriever.Search(ctx, req.Message, 5, req.BaseID)
		if err != nil {
			s.logger.Warn("知识库检索失败", zap.Error(err))
		}
	}

	// 2. 获取MCP工具列表
	var tools []Tool
	if s.mcpSvc != nil {
		var err error
		tools, err = s.mcpSvc.ListTools(ctx)
		if err != nil {
			s.logger.Warn("获取MCP工具列表失败", zap.Error(err))
		}
	}

	// 3. 构建消息历史
	messages := s.buildMessages(req, sources)

	// 4. 调用OpenAI
	response, err := s.openAIClient.Chat(ctx, messages, tools)
	if err != nil {
		return nil, fmt.Errorf("AI聊天失败: %w", err)
	}

	// 5. 处理工具调用
	if len(response.ToolCalls) > 0 && s.mcpSvc != nil {
		for i := range response.ToolCalls {
			toolCall := &response.ToolCalls[i]
			result, err := s.mcpSvc.CallTool(ctx, toolCall.Name, toolCall.Arguments)
			if err != nil {
				s.logger.Error("工具调用失败", zap.Error(err), zap.String("tool", toolCall.Name))
				toolCall.Result = map[string]interface{}{"error": err.Error()}
			} else {
				toolCall.Result = result
			}
		}
	}

	// 添加知识库来源
	response.Sources = sources

	return response, nil
}

// buildMessages 构建消息列表
func (s *Service) buildMessages(req ai.ChatRequest, sources []ai.Source) []ai.ChatMessage {
	messages := []ai.ChatMessage{}

	// 系统提示词
	systemPrompt := s.buildSystemPrompt(sources)
	if systemPrompt != "" {
		messages = append(messages, ai.ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// 历史消息
	messages = append(messages, req.History...)

	// 当前用户消息
	messages = append(messages, ai.ChatMessage{
		Role:    "user",
		Content: req.Message,
	})

	return messages
}

// buildSystemPrompt 构建系统提示词
func (s *Service) buildSystemPrompt(sources []ai.Source) string {
	if len(sources) == 0 {
		return "你是一个友好的AI助手，可以帮助用户解答问题。"
	}

	var builder strings.Builder
	builder.WriteString("你是一个友好的AI助手，可以帮助用户解答问题。\n\n")
	builder.WriteString("以下是相关的知识库内容，请基于这些内容回答用户的问题：\n\n")

	for i, source := range sources {
		builder.WriteString(fmt.Sprintf("【知识%d】%s\n%s\n\n", i+1, source.Title, source.Content))
	}

	builder.WriteString("如果知识库中没有相关信息，请基于你的知识回答，但请说明这是基于通用知识，而非知识库内容。")

	return builder.String()
}

