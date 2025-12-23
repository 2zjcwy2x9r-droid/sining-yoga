package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/domain/knowledge"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// KnowledgeRepository 知识库仓储接口实现
type KnowledgeRepository struct {
	db *gorm.DB
}

// NewKnowledgeRepository 创建知识库仓储
func NewKnowledgeRepository(dsn string) (*KnowledgeRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return &KnowledgeRepository{db: db}, nil
}

// CreateBase 创建知识库
func (r *KnowledgeRepository) CreateBase(ctx context.Context, base *knowledge.KnowledgeBase) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	now := time.Now()
	base.CreatedAt = now
	base.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(base).Error; err != nil {
		return fmt.Errorf("创建知识库失败: %w", err)
	}
	return nil
}

// GetBase 获取知识库
func (r *KnowledgeRepository) GetBase(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeBase, error) {
	var base knowledge.KnowledgeBase
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&base).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识库不存在: %w", err)
		}
		return nil, fmt.Errorf("查询知识库失败: %w", err)
	}
	return &base, nil
}

// ListBases 列出知识库
func (r *KnowledgeRepository) ListBases(ctx context.Context, limit, offset int) ([]*knowledge.KnowledgeBase, error) {
	var bases []*knowledge.KnowledgeBase
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&bases).Error; err != nil {
		return nil, fmt.Errorf("查询知识库列表失败: %w", err)
	}
	return bases, nil
}

// UpdateBase 更新知识库
func (r *KnowledgeRepository) UpdateBase(ctx context.Context, base *knowledge.KnowledgeBase) error {
	base.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(base).Error; err != nil {
		return fmt.Errorf("更新知识库失败: %w", err)
	}
	return nil
}

// DeleteBase 删除知识库
func (r *KnowledgeRepository) DeleteBase(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&knowledge.KnowledgeBase{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除知识库失败: %w", err)
	}
	return nil
}

// CreateItem 创建知识项
func (r *KnowledgeRepository) CreateItem(ctx context.Context, item *knowledge.KnowledgeItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("创建知识项失败: %w", err)
	}
	return nil
}

// GetItem 获取知识项
func (r *KnowledgeRepository) GetItem(ctx context.Context, id uuid.UUID) (*knowledge.KnowledgeItem, error) {
	var item knowledge.KnowledgeItem
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("知识项不存在: %w", err)
		}
		return nil, fmt.Errorf("查询知识项失败: %w", err)
	}
	return &item, nil
}

// ListItems 列出知识项
func (r *KnowledgeRepository) ListItems(ctx context.Context, baseID uuid.UUID, limit, offset int) ([]*knowledge.KnowledgeItem, error) {
	var items []*knowledge.KnowledgeItem
	query := r.db.WithContext(ctx).Where("knowledge_base_id = ?", baseID)
	if err := query.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("查询知识项列表失败: %w", err)
	}
	return items, nil
}

// UpdateItem 更新知识项
func (r *KnowledgeRepository) UpdateItem(ctx context.Context, item *knowledge.KnowledgeItem) error {
	item.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("更新知识项失败: %w", err)
	}
	return nil
}

// UpdateItemEmbeddingStatus 更新知识项的向量化状态
func (r *KnowledgeRepository) UpdateItemEmbeddingStatus(ctx context.Context, id uuid.UUID, status, vectorID string) error {
	updates := map[string]interface{}{
		"embedding_status": status,
		"updated_at":        time.Now(),
	}
	if vectorID != "" {
		updates["vector_id"] = vectorID
	}

	if err := r.db.WithContext(ctx).Model(&knowledge.KnowledgeItem{}).
		Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新向量化状态失败: %w", err)
	}
	return nil
}

// DeleteItem 删除知识项
func (r *KnowledgeRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&knowledge.KnowledgeItem{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除知识项失败: %w", err)
	}
	return nil
}

// GetPendingItems 获取待向量化的知识项
func (r *KnowledgeRepository) GetPendingItems(ctx context.Context, limit int) ([]*knowledge.KnowledgeItem, error) {
	var items []*knowledge.KnowledgeItem
	if err := r.db.WithContext(ctx).
		Where("embedding_status = ?", "pending").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("查询待向量化项失败: %w", err)
	}
	return items, nil
}

