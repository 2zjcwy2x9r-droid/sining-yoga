package knowledge

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/domain/knowledge"
	"github.com/yoga/knowledge-base/pkg/embedding"
	"github.com/yoga/knowledge-base/pkg/observability"
	"github.com/yoga/knowledge-base/pkg/storage"
	"github.com/yoga/knowledge-base/pkg/vector"
	"go.uber.org/zap"
)

// Repository 知识库仓储接口
type Repository interface {
	CreateBase(ctx context.Context, base *knowledge.KnowledgeBase) error
	GetBase(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeBase, error)
	ListBases(ctx context.Context, limit, offset int) ([]*knowledge.KnowledgeBase, error)
	UpdateBase(ctx context.Context, base *knowledge.KnowledgeBase) error
	DeleteBase(ctx context.Context, id uuid.UUID) error
	CreateItem(ctx context.Context, item *knowledge.KnowledgeItem) error
	GetItem(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeItem, error)
	ListItems(ctx context.Context, baseID uuid.UUID, limit, offset int) ([]*knowledge.KnowledgeItem, error)
	UpdateItem(ctx context.Context, item *knowledge.KnowledgeItem) error
	UpdateItemEmbeddingStatus(ctx context.Context, id uuid.UUID, status, vectorID string) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetPendingItems(ctx context.Context, limit int) ([]*knowledge.KnowledgeItem, error)
}

// Service 知识库服务
type Service struct {
	repo         Repository
	storage      storage.Storage
	embeddingSvc *embedding.Client
	vectorSvc    *vector.Client
	bucketName   string
	logger       *zap.Logger
}

// NewService 创建知识库服务
func NewService(repo Repository, storage storage.Storage, embeddingSvc *embedding.Client, vectorSvc *vector.Client, bucketName string, logger *zap.Logger) *Service {
	return &Service{
		repo:         repo,
		storage:      storage,
		embeddingSvc: embeddingSvc,
		vectorSvc:    vectorSvc,
		bucketName:   bucketName,
		logger:       logger,
	}
}

// CreateBase 创建知识库
func (s *Service) CreateBase(ctx context.Context, name, description, baseType string) (*knowledge.KnowledgeBase, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "CreateBase")
	defer span.End()

	base := &knowledge.KnowledgeBase{
		Name:        name,
		Description: description,
		Type:        baseType,
	}

	if err := s.repo.CreateBase(ctx, base); err != nil {
		return nil, fmt.Errorf("创建知识库失败: %w", err)
	}

	s.logger.Info("创建知识库成功", zap.String("id", base.ID.String()))
	return base, nil
}

// GetBase 获取知识库
func (s *Service) GetBase(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeBase, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "GetBase")
	defer span.End()

	return s.repo.GetBase(ctx, id)
}

// ListBases 列出知识库
func (s *Service) ListBases(ctx context.Context, limit, offset int) ([]*knowledge.KnowledgeBase, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "ListBases")
	defer span.End()

	return s.repo.ListBases(ctx, limit, offset)
}

// UpdateBase 更新知识库
func (s *Service) UpdateBase(ctx context.Context, id uuid.UUID, name, description *string) (*knowledge.KnowledgeBase, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "UpdateBase")
	defer span.End()

	base, err := s.repo.GetBase(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		base.Name = *name
	}
	if description != nil {
		base.Description = *description
	}

	if err := s.repo.UpdateBase(ctx, base); err != nil {
		return nil, fmt.Errorf("更新知识库失败: %w", err)
	}

	return base, nil
}

// DeleteBase 删除知识库
func (s *Service) DeleteBase(ctx context.Context, id uuid.UUID) error {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "DeleteBase")
	defer span.End()

	return s.repo.DeleteBase(ctx, id)
}

// CreateTextItem 创建文本知识项
func (s *Service) CreateTextItem(ctx context.Context, baseID uuid.UUID, title, content string) (*knowledge.KnowledgeItem, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "CreateTextItem")
	defer span.End()

	item := &knowledge.KnowledgeItem{
		KnowledgeBaseID: baseID,
		Title:           title,
		Content:         content,
		ContentType:     "text",
		EmbeddingStatus: "pending",
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("创建知识项失败: %w", err)
	}

	// 异步触发向量化
	go s.processEmbedding(context.Background(), item)

	return item, nil
}

// CreateFileItem 创建文件知识项（图片、视频）
func (s *Service) CreateFileItem(ctx context.Context, baseID uuid.UUID, title string, file io.Reader, fileSize int64, contentType, fileName string) (*knowledge.KnowledgeItem, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "CreateFileItem")
	defer span.End()

	// 生成文件路径
	filePath := fmt.Sprintf("%s/%s/%s", baseID.String(), uuid.New().String(), fileName)

	// 上传文件
	if err := s.storage.PutObject(ctx, s.bucketName, filePath, file, fileSize, contentType); err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 确定内容类型
	contentTypeEnum := "image"
	if contentType == "video" || contentType == "video/mp4" || contentType == "video/avi" {
		contentTypeEnum = "video"
	}

	item := &knowledge.KnowledgeItem{
		KnowledgeBaseID: baseID,
		Title:           title,
		ContentType:     contentTypeEnum,
		FilePath:        filePath,
		FileSize:        fileSize,
		MimeType:        contentType,
		EmbeddingStatus: "pending",
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		// 如果创建失败，尝试删除已上传的文件
		_ = s.storage.RemoveObject(ctx, s.bucketName, filePath)
		return nil, fmt.Errorf("创建知识项失败: %w", err)
	}

	// 异步触发向量化（对于图片和视频，可能需要特殊处理）
	go s.processEmbedding(context.Background(), item)

	return item, nil
}

// GetItem 获取知识项
func (s *Service) GetItem(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeItem, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "GetItem")
	defer span.End()

	return s.repo.GetItem(ctx, id)
}

// ListItems 列出知识项
func (s *Service) ListItems(ctx context.Context, baseID uuid.UUID, limit, offset int) ([]*knowledge.KnowledgeItem, error) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "ListItems")
	defer span.End()

	return s.repo.ListItems(ctx, baseID, limit, offset)
}

// DeleteItem 删除知识项
func (s *Service) DeleteItem(ctx context.Context, id uuid.UUID) error {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "DeleteItem")
	defer span.End()

	item, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return err
	}

	// 删除文件
	if item.FilePath != "" {
		if err := s.storage.RemoveObject(ctx, s.bucketName, item.FilePath); err != nil {
			s.logger.Warn("删除文件失败", zap.Error(err), zap.String("file_path", item.FilePath))
		}
	}

	return s.repo.DeleteItem(ctx, id)
}

// processEmbedding 处理向量化
func (s *Service) processEmbedding(ctx context.Context, item *knowledge.KnowledgeItem) {
	ctx, span := observability.StartSpan(ctx, "knowledge-service", "processEmbedding")
	defer span.End()

	// 更新状态为处理中
	if err := s.repo.UpdateItemEmbeddingStatus(ctx, item.ID, "processing", ""); err != nil {
		s.logger.Error("更新向量化状态失败", zap.Error(err))
		return
	}

	// 只处理文本内容
	if item.ContentType != "text" || item.Content == "" {
		s.logger.Info("跳过非文本内容的向量化", zap.String("content_type", item.ContentType))
		if err := s.repo.UpdateItemEmbeddingStatus(ctx, item.ID, "completed", ""); err != nil {
			s.logger.Error("更新向量化状态失败", zap.Error(err))
		}
		return
	}

	// 调用embedding服务
	embeddingVector, err := s.embeddingSvc.EmbedText(ctx, item.Content)
	if err != nil {
		s.logger.Error("向量化失败", zap.Error(err))
		if err := s.repo.UpdateItemEmbeddingStatus(ctx, item.ID, "failed", ""); err != nil {
			s.logger.Error("更新向量化状态失败", zap.Error(err))
		}
		return
	}

	// 存储向量到Qdrant
	vectorID := item.ID.String()
	payload := map[string]interface{}{
		"knowledge_base_id": item.KnowledgeBaseID.String(),
		"item_id":          item.ID.String(),
		"title":            item.Title,
		"content_type":     item.ContentType,
	}

	if err := s.vectorSvc.Store(ctx, vector.StoreRequest{
		ID:      vectorID,
		Vector:  embeddingVector,
		Payload: payload,
	}); err != nil {
		s.logger.Error("存储向量失败", zap.Error(err))
		if err := s.repo.UpdateItemEmbeddingStatus(ctx, item.ID, "failed", ""); err != nil {
			s.logger.Error("更新向量化状态失败", zap.Error(err))
		}
		return
	}

	// 更新状态为完成
	if err := s.repo.UpdateItemEmbeddingStatus(ctx, item.ID, "completed", vectorID); err != nil {
		s.logger.Error("更新向量化状态失败", zap.Error(err))
		return
	}

	s.logger.Info("向量化完成", zap.String("item_id", item.ID.String()), zap.String("vector_id", vectorID))
}

