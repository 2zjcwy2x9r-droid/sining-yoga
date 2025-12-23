package knowledge

import (
	"time"

	"github.com/google/uuid"
)

// KnowledgeBase 知识库实体
type KnowledgeBase struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // 'text', 'image', 'video', 'mixed'
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KnowledgeItem 知识项实体
type KnowledgeItem struct {
	ID              uuid.UUID              `json:"id"`
	KnowledgeBaseID uuid.UUID              `json:"knowledge_base_id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content,omitempty"`
	ContentType     string                 `json:"content_type"` // 'text', 'image', 'video'
	FilePath        string                 `json:"file_path,omitempty"`
	FileSize        int64                  `json:"file_size,omitempty"`
	MimeType        string                 `json:"mime_type,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	VectorID        string                 `json:"vector_id,omitempty"`
	EmbeddingStatus string                 `json:"embedding_status"` // 'pending', 'processing', 'completed', 'failed'
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// IsEmbedded 检查是否已完成向量化
func (ki *KnowledgeItem) IsEmbedded() bool {
	return ki.EmbeddingStatus == "completed" && ki.VectorID != ""
}

