package booking

import (
	"time"

	"github.com/google/uuid"
)

// Class 课程实体
type Class struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Instructor  string    `json:"instructor"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Capacity    int       `json:"capacity"`
	BookedCount int       `json:"booked_count"`
	Status      string    `json:"status"` // 'scheduled', 'cancelled', 'completed'
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Booking 预订实体
type Booking struct {
	ID        uuid.UUID `json:"id"`
	ClassID   uuid.UUID `json:"class_id"`
	UserID    string    `json:"user_id"`    // 微信用户ID
	UserName  string    `json:"user_name"`
	Status    string    `json:"status"`     // 'confirmed', 'cancelled', 'completed'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAvailable 检查课程是否可预订
func (c *Class) IsAvailable() bool {
	return c.Status == "scheduled" && c.BookedCount < c.Capacity
}

// CanBook 检查是否可以预订
func (c *Class) CanBook() bool {
	return c.IsAvailable() && time.Now().Before(c.StartTime)
}

