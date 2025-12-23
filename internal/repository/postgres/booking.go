package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/domain/booking"
	"gorm.io/gorm"
)

// BookingRepository 预订仓储接口实现
type BookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository 创建预订仓储
func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// CreateClass 创建课程
func (r *BookingRepository) CreateClass(ctx context.Context, class *booking.Class) error {
	if class.ID == uuid.Nil {
		class.ID = uuid.New()
	}
	now := time.Now()
	class.CreatedAt = now
	class.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(class).Error; err != nil {
		return fmt.Errorf("创建课程失败: %w", err)
	}
	return nil
}

// GetClass 获取课程
func (r *BookingRepository) GetClass(ctx context.Context, id uuid.UUID) (*booking.Class, error) {
	var class booking.Class
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&class).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("课程不存在: %w", err)
		}
		return nil, fmt.Errorf("查询课程失败: %w", err)
	}
	return &class, nil
}

// ListClasses 列出课程
func (r *BookingRepository) ListClasses(ctx context.Context, startTime, endTime *time.Time, limit, offset int) ([]*booking.Class, error) {
	var classes []*booking.Class
	query := r.db.WithContext(ctx)
	
	if startTime != nil {
		query = query.Where("start_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("end_time <= ?", *endTime)
	}
	
	if err := query.Order("start_time ASC").Limit(limit).Offset(offset).Find(&classes).Error; err != nil {
		return nil, fmt.Errorf("查询课程列表失败: %w", err)
	}
	return classes, nil
}

// UpdateClass 更新课程
func (r *BookingRepository) UpdateClass(ctx context.Context, class *booking.Class) error {
	class.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(class).Error; err != nil {
		return fmt.Errorf("更新课程失败: %w", err)
	}
	return nil
}

// CreateBooking 创建预订
func (r *BookingRepository) CreateBooking(ctx context.Context, b *booking.Booking) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	// 使用事务确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查课程是否存在且可预订
		var class booking.Class
		if err := tx.Where("id = ?", b.ClassID).First(&class).Error; err != nil {
			return fmt.Errorf("课程不存在: %w", err)
		}

		if !class.CanBook() {
			return fmt.Errorf("课程不可预订")
		}

		// 检查是否已预订
		var existing booking.Booking
		if err := tx.Where("class_id = ? AND user_id = ? AND status = ?", b.ClassID, b.UserID, "confirmed").First(&existing).Error; err == nil {
			return fmt.Errorf("您已预订该课程")
		}

		// 创建预订
		if err := tx.Create(b).Error; err != nil {
			return fmt.Errorf("创建预订失败: %w", err)
		}

		// 更新课程预订数量
		if err := tx.Model(&class).Update("booked_count", gorm.Expr("booked_count + 1")).Error; err != nil {
			return fmt.Errorf("更新课程预订数量失败: %w", err)
		}

		return nil
	})
}

// GetBooking 获取预订
func (r *BookingRepository) GetBooking(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	var b booking.Booking
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("预订不存在: %w", err)
		}
		return nil, fmt.Errorf("查询预订失败: %w", err)
	}
	return &b, nil
}

// ListUserBookings 列出用户预订
func (r *BookingRepository) ListUserBookings(ctx context.Context, userID string, limit, offset int) ([]*booking.Booking, error) {
	var bookings []*booking.Booking
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("查询用户预订列表失败: %w", err)
	}
	return bookings, nil
}

// CancelBooking 取消预订
func (r *BookingRepository) CancelBooking(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b booking.Booking
		if err := tx.Where("id = ?", id).First(&b).Error; err != nil {
			return fmt.Errorf("预订不存在: %w", err)
		}

		if b.Status != "confirmed" {
			return fmt.Errorf("只能取消已确认的预订")
		}

		// 更新预订状态
		if err := tx.Model(&b).Update("status", "cancelled").Error; err != nil {
			return fmt.Errorf("取消预订失败: %w", err)
		}

		// 更新课程预订数量
		if err := tx.Model(&booking.Class{}).
			Where("id = ?", b.ClassID).
			Update("booked_count", gorm.Expr("booked_count - 1")).Error; err != nil {
			return fmt.Errorf("更新课程预订数量失败: %w", err)
		}

		return nil
	})
}

// CreateReview 创建评价
func (r *BookingRepository) CreateReview(ctx context.Context, review *booking.Review) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("创建评价失败: %w", err)
	}
	return nil
}

// ListClassReviews 列出课程评价
func (r *BookingRepository) ListClassReviews(ctx context.Context, classID uuid.UUID, limit, offset int) ([]*booking.Review, error) {
	var reviews []*booking.Review
	if err := r.db.WithContext(ctx).
		Where("class_id = ?", classID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("查询课程评价列表失败: %w", err)
	}
	return reviews, nil
}

// GetUserReview 获取用户对课程的评价
func (r *BookingRepository) GetUserReview(ctx context.Context, classID uuid.UUID, userID string) (*booking.Review, error) {
	var review booking.Review
	if err := r.db.WithContext(ctx).
		Where("class_id = ? AND user_id = ?", classID, userID).
		First(&review).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户评价失败: %w", err)
	}
	return &review, nil
}

