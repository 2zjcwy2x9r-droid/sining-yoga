package booking

import (
	"time"

	"github.com/google/uuid"
)

// Review 课程评价实体
type Review struct {
	ID        uuid.UUID `json:"id"`
	ClassID   uuid.UUID `json:"class_id"`
	UserID    string    `json:"user_id"`    // 微信用户ID
	UserName  string    `json:"user_name"`  // 用户名称
	Rating    int       `json:"rating"`     // 评分 1-5
	Content   string    `json:"content"`    // 评价内容
	Images    []string  `json:"images"`     // 评价图片URL列表
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsValidRating 检查评分是否有效
func (r *Review) IsValidRating() bool {
	return r.Rating >= 1 && r.Rating <= 5
}

