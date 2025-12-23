package ai

import (
	"context"
	"fmt"

	"github.com/yoga/knowledge-base/internal/domain/ai"
	"github.com/yoga/knowledge-base/internal/repository/postgres"
	"github.com/yoga/knowledge-base/pkg/vector"
	"go.uber.org/zap"
)

// KnowledgeRetriever 知识库检索器实现
type KnowledgeRetriever struct {
	vectorClient *vector.Client
	kbRepo       *postgres.KnowledgeRepository
	logger       *zap.Logger
}

// NewRetriever 创建检索器
func NewRetriever(vectorClient *vector.Client, kbRepo *postgres.KnowledgeRepository, logger *zap.Logger) Retriever {
	return &KnowledgeRetriever{
		vectorClient: vectorClient,
		kbRepo:       kbRepo,
		logger:       logger,
	}
}

// Search 检索知识库内容
func (r *KnowledgeRetriever) Search(ctx context.Context, query string, limit int, knowledgeBaseID string) ([]ai.Source, error) {
	// 调用向量服务检索
	searchReq := vector.SearchRequest{
		Query:           query,
		Limit:           limit,
		KnowledgeBaseID: knowledgeBaseID,
	}

	searchResp, err := r.vectorClient.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	// 转换为Source
	sources := make([]ai.Source, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		// 从payload中提取信息
		title, _ := result.Payload["title"].(string)
		itemID, _ := result.Payload["item_id"].(string)

		// 如果需要，可以从数据库获取完整内容
		var content string
		if itemID != "" {
			// TODO: 从数据库获取完整内容
			content = title // 暂时使用标题
		}

		sources = append(sources, ai.Source{
			ID:      result.ID,
			Title:   title,
			Content: content,
			Score:   result.Score,
		})
	}

	return sources, nil
}
